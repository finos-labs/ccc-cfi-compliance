package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AWSIAMService implements IAMService for AWS
type AWSIAMService struct {
	client           *iam.Client
	ctx              context.Context
	provisionedUsers map[string]*Identity // Cache of provisioned users by userName
	accessLevels     map[string]string    // Cache of access levels by "userName:serviceID"
}

// NewAWSIAMService creates a new AWS IAM service using default credentials
func NewAWSIAMService(ctx context.Context) (*AWSIAMService, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSIAMService{
		client:           iam.NewFromConfig(cfg),
		ctx:              ctx,
		provisionedUsers: make(map[string]*Identity),
		accessLevels:     make(map[string]string),
	}, nil
}

// ProvisionUser creates a new IAM user with access keys
func (s *AWSIAMService) ProvisionUser(userName string) (*Identity, error) {
	// Check cache first - if we've already provisioned this user in this session, return it
	if cachedIdentity, exists := s.provisionedUsers[userName]; exists {
		fmt.Printf("♻️  Using cached identity for user %s (skipping propagation delay)\n", userName)
		return cachedIdentity, nil
	}

	var createUserOutput *iam.CreateUserOutput
	var userAlreadyExists bool

	// Check if user already exists
	getUserOutput, err := s.client.GetUser(s.ctx, &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err == nil {
		// User exists - reuse it
		fmt.Printf("👤 User %s already exists, reusing...\n", userName)
		createUserOutput = &iam.CreateUserOutput{User: getUserOutput.User}
		userAlreadyExists = true
	} else {
		// User doesn't exist - create it
		fmt.Printf("👤 Creating user %s...\n", userName)
		createUserOutput, err = s.client.CreateUser(s.ctx, &iam.CreateUserInput{
			UserName: aws.String(userName),
			Tags: []types.Tag{
				{
					Key:   aws.String("Purpose"),
					Value: aws.String("CCC-Testing"),
				},
				{
					Key:   aws.String("ManagedBy"),
					Value: aws.String("CCC-CFI-Compliance-Framework"),
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create IAM user %s: %w", userName, err)
		}
	}

	// Create access key for the user
	// Note: We always create a new access key because we can't retrieve secrets for existing keys
	var accessKeyId, secretAccessKey string

	// If user already exists, delete any existing access keys to avoid hitting the limit (AWS allows max 2 keys)
	if userAlreadyExists {
		listKeysOutput, err := s.client.ListAccessKeys(s.ctx, &iam.ListAccessKeysInput{
			UserName: aws.String(userName),
		})
		if err == nil {
			for _, keyMetadata := range listKeysOutput.AccessKeyMetadata {
				fmt.Printf("   🗑️  Deleting old access key: %s\n", aws.ToString(keyMetadata.AccessKeyId))
				_, err := s.client.DeleteAccessKey(s.ctx, &iam.DeleteAccessKeyInput{
					UserName:    aws.String(userName),
					AccessKeyId: keyMetadata.AccessKeyId,
				})
				if err != nil {
					fmt.Printf("   ⚠️  Failed to delete old access key %s: %v\n", aws.ToString(keyMetadata.AccessKeyId), err)
				}
			}
		}
	}

	// Create new access key
	createKeyOutput, err := s.client.CreateAccessKey(s.ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		// Cleanup: delete the user if key creation fails (only if we just created it)
		if !userAlreadyExists {
			s.client.DeleteUser(s.ctx, &iam.DeleteUserInput{
				UserName: aws.String(userName),
			})
		}
		return nil, fmt.Errorf("failed to create access key for user %s: %w", userName, err)
	}
	accessKeyId = aws.ToString(createKeyOutput.AccessKey.AccessKeyId)
	secretAccessKey = aws.ToString(createKeyOutput.AccessKey.SecretAccessKey)
	fmt.Printf("   🔑 Created new access key: %s\n", accessKeyId)

	// Create identity with credentials in map
	identity := &Identity{
		UserName:    userName,
		Provider:    "aws",
		Credentials: make(map[string]string),
	}

	// Store AWS-specific fields in Credentials map
	identity.Credentials["arn"] = aws.ToString(createUserOutput.User.Arn)
	identity.Credentials["user_id"] = aws.ToString(createUserOutput.User.UserId)
	identity.Credentials["access_key_id"] = accessKeyId
	if secretAccessKey != "" {
		identity.Credentials["secret_access_key"] = secretAccessKey
	}

	// Extract and store account ID from ARN (format: arn:aws:iam::123456789012:user/username)
	if createUserOutput.User.Arn != nil {
		arn := aws.ToString(createUserOutput.User.Arn)
		parts := splitARN(arn)
		if len(parts) > 4 {
			identity.Credentials["account_id"] = parts[4]
		}
	}

	// Log the created/retrieved identity details
	fmt.Printf("✅ Provisioned user: %s\n", userName)
	fmt.Printf("   ARN: %s\n", identity.Credentials["arn"])
	fmt.Printf("   User ID: %s\n", identity.Credentials["user_id"])
	fmt.Printf("   Access Key: %s\n", identity.Credentials["access_key_id"])
	fmt.Printf("   Secret Key Length: %d\n", len(identity.Credentials["secret_access_key"]))
	if identity.Credentials["account_id"] != "" {
		fmt.Printf("   Account ID: %s\n", identity.Credentials["account_id"])
	}

	// Cache the identity for future requests
	s.provisionedUsers[userName] = identity

	return identity, nil
}

// SetAccess grants an identity access to a specific AWS service/resource at the specified level
func (s *AWSIAMService) SetAccess(identity *Identity, serviceID string, level string) (string, error) {
	// Check cache first - if we've already set this access level, skip it
	cacheKey := fmt.Sprintf("%s:%s", identity.UserName, serviceID)
	if cachedLevel, exists := s.accessLevels[cacheKey]; exists && cachedLevel == level {
		fmt.Printf("♻️  Access level already set to %s for %s (skipping propagation delay)\n", level, identity.UserName)
		// Generate and return the policy document without making AWS calls
		policyDocument, err := s.generatePolicyDocument(serviceID, level)
		if err != nil {
			return "", fmt.Errorf("failed to generate policy: %w", err)
		}
		return policyDocument, nil
	}

	// Check current access level
	currentLevel, currentPolicy, err := s.GetAccess(identity, serviceID)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not retrieve current access level: %v\n", err)
	} else {
		fmt.Printf("📊 Current access level: %s → New access level: %s\n", currentLevel, level)

		if currentLevel == level {
			fmt.Printf("ℹ️  Access level unchanged, skipping...\n")
			// Cache this access level
			s.accessLevels[cacheKey] = level
			// Return the existing policy document
			return currentPolicy, nil
		}
	}

	// Generate policy document based on access level and service ID
	policyDocument, err := s.generatePolicyDocument(serviceID, level)
	if err != nil {
		return "", fmt.Errorf("failed to generate policy: %w", err)
	}

	// Create a unique policy name
	policyName := fmt.Sprintf("CCC-Test-%s-%s", sanitizeForPolicyName(serviceID), level)

	// Attach inline policy to user
	_, err = s.client.PutUserPolicy(s.ctx, &iam.PutUserPolicyInput{
		UserName:       aws.String(identity.UserName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach policy to user %s: %w", identity.UserName, err)
	}

	fmt.Printf("📋 Attached policy '%s' to user %s\n", policyName, identity.UserName)

	// Wait for IAM policy propagation
	fmt.Printf("⏳ Waiting 15 seconds for IAM policy changes to propagate...\n")
	time.Sleep(15 * time.Second)
	fmt.Printf("✅ IAM policy propagation wait complete\n")

	// Cache the access level for future requests
	s.accessLevels[cacheKey] = level

	return policyDocument, nil
}

// GetAccess retrieves the current access level for a user and service
func (s *AWSIAMService) GetAccess(identity *Identity, serviceID string) (string, string, error) {
	// List all inline policies for the user
	listPoliciesOutput, err := s.client.ListUserPolicies(s.ctx, &iam.ListUserPoliciesInput{
		UserName: aws.String(identity.UserName),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to list user policies: %w", err)
	}

	// Look for the specific policy we manage
	policyPrefix := fmt.Sprintf("CCC-Test-%s-", sanitizeForPolicyName(serviceID))

	for _, policyName := range listPoliciesOutput.PolicyNames {
		// Check if this policy matches our service
		if len(policyName) >= len(policyPrefix) && policyName[:len(policyPrefix)] == policyPrefix {
			// Get the policy document
			getPolicyOutput, err := s.client.GetUserPolicy(s.ctx, &iam.GetUserPolicyInput{
				UserName:   aws.String(identity.UserName),
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				return "", "", fmt.Errorf("failed to get policy %s: %w", policyName, err)
			}

			// Extract access level from policy name
			// Policy name format: "CCC-Test-{serviceID}-{level}"
			level := policyName[len(policyPrefix):]

			// Get the policy document (it's URL-encoded in the response)
			policyDocument := aws.ToString(getPolicyOutput.PolicyDocument)

			// Decode the URL-encoded policy document
			decodedPolicy, err := url.QueryUnescape(policyDocument)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to decode policy document: %v\n", err)
				decodedPolicy = policyDocument
			}

			fmt.Printf("📋 Current policy '%s' grants '%s' access\n", policyName, level)

			return level, decodedPolicy, nil
		}
	}

	// No matching policy found
	return "none", "", nil
}

// DestroyUser removes an IAM user and all associated resources
func (s *AWSIAMService) DestroyUser(identity *Identity) error {
	userName := identity.UserName

	// List and delete access keys
	listKeysOutput, err := s.client.ListAccessKeys(s.ctx, &iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to list access keys for user %s: %w", userName, err)
	}

	for _, key := range listKeysOutput.AccessKeyMetadata {
		_, err := s.client.DeleteAccessKey(s.ctx, &iam.DeleteAccessKeyInput{
			UserName:    aws.String(userName),
			AccessKeyId: key.AccessKeyId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete access key %s: %w", aws.ToString(key.AccessKeyId), err)
		}
	}

	// List and delete inline policies
	listPoliciesOutput, err := s.client.ListUserPolicies(s.ctx, &iam.ListUserPoliciesInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to list user policies for %s: %w", userName, err)
	}

	for _, policyName := range listPoliciesOutput.PolicyNames {
		_, err := s.client.DeleteUserPolicy(s.ctx, &iam.DeleteUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete policy %s: %w", policyName, err)
		}
	}

	// List and detach managed policies
	listAttachedOutput, err := s.client.ListAttachedUserPolicies(s.ctx, &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to list attached policies for %s: %w", userName, err)
	}

	for _, policy := range listAttachedOutput.AttachedPolicies {
		_, err := s.client.DetachUserPolicy(s.ctx, &iam.DetachUserPolicyInput{
			UserName:  aws.String(userName),
			PolicyArn: policy.PolicyArn,
		})
		if err != nil {
			return fmt.Errorf("failed to detach policy %s: %w", aws.ToString(policy.PolicyArn), err)
		}
	}

	// Finally, delete the user
	_, err = s.client.DeleteUser(s.ctx, &iam.DeleteUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", userName, err)
	}

	return nil
}

// generatePolicyDocument creates an IAM policy document for the given resource and access level
func (s *AWSIAMService) generatePolicyDocument(resourceIdentifier string, level string) (string, error) {
	var statements []map[string]interface{}

	// Determine actions based on service type and access level
	switch level {
	case "none":
		// No permissions granted - return empty policy
		policy := map[string]interface{}{
			"Version":   "2012-10-17",
			"Statement": []map[string]interface{}{},
		}
		policyJSON, err := json.Marshal(policy)
		if err != nil {
			return "", fmt.Errorf("failed to marshal policy: %w", err)
		}
		return string(policyJSON), nil

	case "read":
		// S3 bucket-level permissions (ListBucket requires the bucket ARN)
		statements = append(statements, map[string]interface{}{
			"Effect": "Allow",
			"Action": []string{
				"s3:ListBucket",
				"s3:GetBucketLocation",
			},
			"Resource": fmt.Sprintf("arn:aws:s3:::%s", resourceIdentifier),
		})
		// S3 object-level permissions (GetObject requires bucket/* ARN)
		statements = append(statements, map[string]interface{}{
			"Effect": "Allow",
			"Action": []string{
				"s3:GetObject",
			},
			"Resource": fmt.Sprintf("arn:aws:s3:::%s/*", resourceIdentifier),
		})

	case "write":
		// S3 bucket-level permissions
		statements = append(statements, map[string]interface{}{
			"Effect": "Allow",
			"Action": []string{
				"s3:ListBucket",
				"s3:GetBucketLocation",
			},
			"Resource": fmt.Sprintf("arn:aws:s3:::%s", resourceIdentifier),
		})
		// S3 object-level permissions
		statements = append(statements, map[string]interface{}{
			"Effect": "Allow",
			"Action": []string{
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
			},
			"Resource": fmt.Sprintf("arn:aws:s3:::%s/*", resourceIdentifier),
		})

	case "admin":
		// Full S3 permissions
		statements = append(statements, map[string]interface{}{
			"Effect":   "Allow",
			"Action":   "*",
			"Resource": "*",
		})

	default:
		return "", fmt.Errorf("unsupported access level: %s", level)
	}

	// Build policy document
	policy := map[string]interface{}{
		"Version":   "2012-10-17",
		"Statement": statements,
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy: %w", err)
	}

	return string(policyJSON), nil
}

// Helper functions

func splitARN(arn string) []string {
	// Simple ARN splitter: arn:partition:service:region:account-id:resource
	result := make([]string, 0)
	current := ""
	for _, char := range arn {
		if char == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func sanitizeForPolicyName(s string) string {
	// Replace characters that aren't valid in policy names
	result := ""
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else if char == '-' || char == '_' {
			result += string(char)
		}
	}
	if len(result) > 64 {
		result = result[:64]
	}
	return result
}

// Fill this later when we are writing tests for IAM
func (s *AWSIAMService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	return []environment.TestParams{}, nil
}

// ElevateAccessForInspection is a no-op for IAM services
func (s *AWSIAMService) ElevateAccessForInspection() error {
	// No-op: IAM services don't have network-level access controls to elevate
	return nil
}

// ResetAccess is a no-op for IAM services
func (s *AWSIAMService) ResetAccess() error {
	// No-op: IAM services don't have network-level access controls to reset
	return nil
}
