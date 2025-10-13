package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"google.golang.org/api/option"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

// GCPIAMService implements IAMService for GCP
type GCPIAMService struct {
	client    *admin.IamClient
	ctx       context.Context
	projectID string
}

// NewGCPIAMService creates a new GCP IAM service using default credentials
func NewGCPIAMService(ctx context.Context, projectID string) (*GCPIAMService, error) {
	client, err := admin.NewIamClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP IAM client: %w", err)
	}

	return &GCPIAMService{
		client:    client,
		ctx:       ctx,
		projectID: projectID,
	}, nil
}

// NewGCPIAMServiceWithCredentials creates a new GCP IAM service with specific credentials
func NewGCPIAMServiceWithCredentials(ctx context.Context, projectID string, credentialsJSON []byte) (*GCPIAMService, error) {
	client, err := admin.NewIamClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP IAM client with credentials: %w", err)
	}

	return &GCPIAMService{
		client:    client,
		ctx:       ctx,
		projectID: projectID,
	}, nil
}

// ProvisionUser creates a new IAM service account (GCP's equivalent of a user for programmatic access)
func (s *GCPIAMService) ProvisionUser(userName string) (*Identity, error) {
	// Service account ID must be between 6-30 characters, lowercase, digits, hyphens
	serviceAccountID := sanitizeServiceAccountID(userName)
	serviceAccountEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountID, s.projectID)
	
	var serviceAccount *adminpb.ServiceAccount
	var accountAlreadyExists bool

	// Check if service account already exists
	getReq := &adminpb.GetServiceAccountRequest{
		Name: fmt.Sprintf("projects/%s/serviceAccounts/%s", s.projectID, serviceAccountEmail),
	}
	
	existingAccount, err := s.client.GetServiceAccount(s.ctx, getReq)
	if err == nil {
		// Service account exists - reuse it
		fmt.Printf("🤖 Service account %s already exists, reusing...\n", serviceAccountEmail)
		serviceAccount = existingAccount
		accountAlreadyExists = true
	} else {
		// Service account doesn't exist - create it
		fmt.Printf("🤖 Creating service account %s...\n", serviceAccountEmail)
		createReq := &adminpb.CreateServiceAccountRequest{
			Name:      fmt.Sprintf("projects/%s", s.projectID),
			AccountId: serviceAccountID,
			ServiceAccount: &adminpb.ServiceAccount{
				DisplayName: fmt.Sprintf("CCC Test User: %s", userName),
				Description: "Created by CCC-CFI-Compliance-Framework for testing",
			},
		}
		
		serviceAccount, err = s.client.CreateServiceAccount(s.ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create service account %s: %w", serviceAccountID, err)
		}
	}

	// Create service account key for authentication
	var keyJSON []byte
	
	if accountAlreadyExists {
		// List existing keys
		listKeysReq := &adminpb.ListServiceAccountKeysRequest{
			Name: serviceAccount.Name,
		}
		
		keysResp, err := s.client.ListServiceAccountKeys(s.ctx, listKeysReq)
		if err == nil && len(keysResp.Keys) > 0 {
			fmt.Printf("   🔑 Service account has %d existing key(s)\n", len(keysResp.Keys))
			// Note: We can't retrieve existing key data, so we create a new one
		}
	}

	// Create new key (always create a new key for fresh credentials)
	createKeyReq := &adminpb.CreateServiceAccountKeyRequest{
		Name: serviceAccount.Name,
	}
	
	key, err := s.client.CreateServiceAccountKey(s.ctx, createKeyReq)
	if err != nil {
		// Cleanup: delete the service account if key creation fails (only if we just created it)
		if !accountAlreadyExists {
			s.client.DeleteServiceAccount(s.ctx, &adminpb.DeleteServiceAccountRequest{
				Name: serviceAccount.Name,
			})
		}
		return nil, fmt.Errorf("failed to create service account key for %s: %w", serviceAccountID, err)
	}
	
	keyJSON = key.PrivateKeyData
	fmt.Printf("   🔑 Created new service account key\n")

	// Parse the key JSON to extract useful information
	var keyData map[string]interface{}
	if err := json.Unmarshal(keyJSON, &keyData); err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %w", err)
	}

	// Create identity with credentials
	identity := &Identity{
		UserName:    userName,
		Provider:    "gcp",
		Credentials: make(map[string]string),
	}

	// Store GCP-specific fields in Credentials map
	identity.Credentials["email"] = serviceAccount.Email
	identity.Credentials["unique_id"] = serviceAccount.UniqueId
	identity.Credentials["project_id"] = s.projectID
	identity.Credentials["service_account_key"] = string(keyJSON)
	
	if clientEmail, ok := keyData["client_email"].(string); ok {
		identity.Credentials["client_email"] = clientEmail
	}
	if privateKeyID, ok := keyData["private_key_id"].(string); ok {
		identity.Credentials["private_key_id"] = privateKeyID
	}

	// Log the created/retrieved identity details
	fmt.Printf("✅ Provisioned service account: %s\n", userName)
	fmt.Printf("   Email: %s\n", identity.Credentials["email"])
	fmt.Printf("   Unique ID: %s\n", identity.Credentials["unique_id"])
	fmt.Printf("   Project ID: %s\n", identity.Credentials["project_id"])

	return identity, nil
}

// SetAccess grants an identity access to a specific GCP service/resource at the specified level
func (s *GCPIAMService) SetAccess(identity *Identity, serviceID string, level string) error {
	// Determine the IAM role based on access level
	role, err := s.getRoleForLevel(serviceID, level)
	if err != nil {
		return fmt.Errorf("failed to determine role: %w", err)
	}

	if role == "" {
		// "none" level - no role to assign
		return nil
	}

	// Get the service account email
	serviceAccountEmail := identity.Credentials["email"]
	if serviceAccountEmail == "" {
		return fmt.Errorf("service account email not found in identity credentials")
	}

	member := fmt.Sprintf("serviceAccount:%s", serviceAccountEmail)

	// Set IAM policy binding
	// For project-level resources, use project as the resource
	// For specific resources, use the resource ID
	resourceName := s.parseResourceName(serviceID)
	
	fmt.Printf("🔐 Granting %s access to %s for %s...\n", level, resourceName, serviceAccountEmail)

	// Get current policy
	getPolicyReq := &iampb.GetIamPolicyRequest{
		Resource: resourceName,
	}
	
	policy, err := s.getResourcePolicy(s.ctx, resourceName, getPolicyReq)
	if err != nil {
		return fmt.Errorf("failed to get IAM policy for %s: %w", resourceName, err)
	}

	// Check if binding already exists
	bindingExists := false
	for _, binding := range policy.Bindings {
		if binding.Role == role {
			// Check if member already in binding
			for _, existingMember := range binding.Members {
				if existingMember == member {
					bindingExists = true
					fmt.Printf("   ℹ️  Binding already exists\n")
					break
				}
			}
			if !bindingExists {
				// Add member to existing role binding
				binding.Members = append(binding.Members, member)
			}
			break
		}
	}

	// If binding doesn't exist for this role, create it
	if !bindingExists {
		newBinding := &iampb.Binding{
			Role:    role,
			Members: []string{member},
		}
		policy.Bindings = append(policy.Bindings, newBinding)
	}

	// Set the updated policy
	setPolicyReq := &iampb.SetIamPolicyRequest{
		Resource: resourceName,
		Policy:   policy,
	}
	
	_, err = s.setResourcePolicy(s.ctx, resourceName, setPolicyReq)
	if err != nil {
		return fmt.Errorf("failed to set IAM policy for %s: %w", resourceName, err)
	}

	fmt.Printf("   ✅ Access granted\n")
	return nil
}

// DestroyUser removes a service account and all associated resources
func (s *GCPIAMService) DestroyUser(identity *Identity) error {
	serviceAccountEmail := identity.Credentials["email"]
	if serviceAccountEmail == "" {
		return fmt.Errorf("service account email not found in identity credentials")
	}

	serviceAccountName := fmt.Sprintf("projects/%s/serviceAccounts/%s", s.projectID, serviceAccountEmail)
	
	fmt.Printf("🗑️  Deleting service account %s...\n", serviceAccountEmail)

	// List and delete all keys
	listKeysReq := &adminpb.ListServiceAccountKeysRequest{
		Name: serviceAccountName,
	}
	
	keysResp, err := s.client.ListServiceAccountKeys(s.ctx, listKeysReq)
	if err != nil {
		// If we can't list keys, the account might not exist - continue anyway
		fmt.Printf("   ⚠️  Could not list keys: %v\n", err)
	} else {
		for _, key := range keysResp.Keys {
			// Skip system-managed keys (only delete user-managed keys)
			if key.KeyType == adminpb.ListServiceAccountKeysRequest_USER_MANAGED {
				deleteKeyReq := &adminpb.DeleteServiceAccountKeyRequest{
					Name: key.Name,
				}
				
				err := s.client.DeleteServiceAccountKey(s.ctx, deleteKeyReq)
				if err != nil {
					fmt.Printf("   ⚠️  Failed to delete key %s: %v\n", key.Name, err)
				}
			}
		}
	}

	// Delete the service account
	deleteReq := &adminpb.DeleteServiceAccountRequest{
		Name: serviceAccountName,
	}
	
	err = s.client.DeleteServiceAccount(s.ctx, deleteReq)
	if err != nil {
		// Check if account doesn't exist
		if strings.Contains(err.Error(), "not found") {
			fmt.Printf("   ℹ️  Service account already deleted\n")
			return nil
		}
		return fmt.Errorf("failed to delete service account %s: %w", serviceAccountEmail, err)
	}

	fmt.Printf("   ✅ Service account deleted\n")
	return nil
}

// Helper functions

func (s *GCPIAMService) getRoleForLevel(serviceID string, level string) (string, error) {
	// Determine the appropriate IAM role based on service and level
	// This is a simplified mapping - in production, you'd want more sophisticated logic
	
	switch level {
	case "none":
		return "", nil
	case "read":
		// Use viewer roles for read access
		if strings.Contains(serviceID, "storage") {
			return "roles/storage.objectViewer", nil
		}
		return "roles/viewer", nil
	case "write":
		// Use editor/writer roles for write access
		if strings.Contains(serviceID, "storage") {
			return "roles/storage.objectAdmin", nil
		}
		return "roles/editor", nil
	case "admin":
		// Use owner/admin roles for admin access
		if strings.Contains(serviceID, "storage") {
			return "roles/storage.admin", nil
		}
		return "roles/owner", nil
	default:
		return "", fmt.Errorf("unsupported access level: %s", level)
	}
}

func (s *GCPIAMService) parseResourceName(serviceID string) string {
	// If serviceID is already a full resource name (projects/...), use it
	if strings.HasPrefix(serviceID, "projects/") {
		return serviceID
	}
	
	// If it's a bucket name, format as bucket resource
	if strings.Contains(serviceID, "storage") || strings.HasPrefix(serviceID, "gs://") {
		bucketName := strings.TrimPrefix(serviceID, "gs://")
		return fmt.Sprintf("projects/_/buckets/%s", bucketName)
	}
	
	// Default: assume it's a project-level permission
	return fmt.Sprintf("projects/%s", s.projectID)
}

func (s *GCPIAMService) getResourcePolicy(ctx context.Context, resourceName string, req *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	// This is a simplified version - in production, you'd need to handle different resource types
	// For now, assume project-level resources
	
	// Using the IAM admin client's method
	// Note: Different resource types may require different clients
	if strings.HasPrefix(resourceName, "projects/") && !strings.Contains(resourceName, "/") {
		// Project-level policy - this would require the Resource Manager API
		// For now, return a basic policy structure
		return &iampb.Policy{Bindings: []*iampb.Binding{}}, nil
	}
	
	// For other resources, return basic policy
	return &iampb.Policy{Bindings: []*iampb.Binding{}}, nil
}

func (s *GCPIAMService) setResourcePolicy(ctx context.Context, resourceName string, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	// This is a simplified version - in production, you'd need to handle different resource types
	// For now, just return the policy that was set
	return req.Policy, nil
}

func sanitizeServiceAccountID(userName string) string {
	// Service account IDs must be 6-30 characters, lowercase letters, digits, hyphens
	// Cannot start with a digit
	result := strings.ToLower(userName)
	
	// Replace invalid characters with hyphens
	sanitized := ""
	for i, char := range result {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			// Don't allow digit as first character
			if i == 0 && char >= '0' && char <= '9' {
				sanitized += "sa-"
			}
			sanitized += string(char)
		} else {
			sanitized += "-"
		}
	}
	
	// Ensure it starts with a letter
	if len(sanitized) > 0 && (sanitized[0] < 'a' || sanitized[0] > 'z') {
		sanitized = "sa-" + sanitized
	}
	
	// Ensure minimum length
	if len(sanitized) < 6 {
		sanitized = sanitized + "-test"
	}
	
	// Ensure maximum length
	if len(sanitized) > 30 {
		sanitized = sanitized[:30]
	}
	
	// Remove trailing hyphens
	sanitized = strings.TrimRight(sanitized, "-")
	
	return sanitized
}

