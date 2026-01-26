package objstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AWSS3Service implements Service for AWS S3
type AWSS3Service struct {
	client      *s3.Client
	config      aws.Config
	ctx         context.Context
	cloudParams environment.CloudParams
}

// NewAWSS3Service creates a new AWS S3 service using default credentials
func NewAWSS3Service(ctx context.Context, cloudParams environment.CloudParams) (*AWSS3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cloudParams.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSS3Service{
		client:      s3.NewFromConfig(cfg),
		config:      cfg,
		ctx:         ctx,
		cloudParams: cloudParams,
	}, nil
}

// NewAWSS3ServiceWithCredentials creates a new AWS S3 service with specific credentials from an Identity
func NewAWSS3ServiceWithCredentials(ctx context.Context, cloudParams environment.CloudParams, identity *iam.Identity) (*AWSS3Service, error) {
	// Extract credentials from the map
	accessKeyID := identity.Credentials["access_key_id"]
	secretAccessKey := identity.Credentials["secret_access_key"]
	sessionToken := identity.Credentials["session_token"] // Optional, empty string if not present

	// Debug logging
	fmt.Printf("🔐 Creating S3 client with credentials:\n")
	fmt.Printf("   Access Key ID: %s\n", accessKeyID)
	fmt.Printf("   Secret Key Length: %d\n", len(secretAccessKey))
	fmt.Printf("   Has Session Token: %v\n", sessionToken != "")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cloudParams.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			sessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config with credentials: %w", err)
	}

	return &AWSS3Service{
		client:      s3.NewFromConfig(cfg),
		config:      cfg,
		ctx:         ctx,
		cloudParams: cloudParams,
	}, nil
}

// ListBuckets lists all S3 buckets
func (s *AWSS3Service) ListBuckets() ([]Bucket, error) {
	output, err := s.client.ListBuckets(s.ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]Bucket, 0, len(output.Buckets))
	for _, b := range output.Buckets {
		bucketName := aws.ToString(b.Name)

		// Get the region for this bucket
		region, err := s.GetBucketRegion(bucketName)
		if err != nil {
			// If we can't get the region, log a warning but continue
			fmt.Printf("⚠️  Warning: Failed to get region for bucket %s: %v\n", bucketName, err)
			region = ""
		}

		buckets = append(buckets, Bucket{
			ID:     bucketName,
			Name:   bucketName,
			Region: region,
		})
	}

	return buckets, nil
}

// CreateBucket creates a new S3 bucket in the configured region
func (s *AWSS3Service) CreateBucket(bucketID string) (*Bucket, error) {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketID),
	}

	_, err := regionalClient.CreateBucket(s.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket %s: %w", bucketID, err)
	}

	return &Bucket{
		ID:     bucketID,
		Name:   bucketID,
		Region: s.cloudParams.Region,
	}, nil
}

// DeleteBucket deletes an S3 bucket
func (s *AWSS3Service) DeleteBucket(bucketID string) error {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	_, err := regionalClient.DeleteBucket(s.ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %w", bucketID, err)
	}

	return nil
}

// ListObjects lists all objects in a bucket
func (s *AWSS3Service) ListObjects(bucketID string) ([]Object, error) {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	output, err := regionalClient.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s: %w", bucketID, err)
	}

	objects := make([]Object, 0, len(output.Contents))
	for _, obj := range output.Contents {
		objects = append(objects, Object{
			ID:       aws.ToString(obj.Key),
			BucketID: bucketID,
			Name:     aws.ToString(obj.Key),
			Size:     aws.ToInt64(obj.Size),
			Data:     nil, // Don't fetch data in list operation
		})
	}

	return objects, nil
}

// CreateObject creates a new object in a bucket
func (s *AWSS3Service) CreateObject(bucketID string, objectID string, data string) (*Object, error) {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	// Convert string to []byte
	content := []byte(data)

	_, err := regionalClient.PutObject(s.ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
		Body:   bytes.NewReader(content),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create object %s in bucket %s: %w", objectID, bucketID, err)
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     int64(len(content)),
		Data:     []string{data},
	}, nil
}

// ReadObject reads an object from a bucket
func (s *AWSS3Service) ReadObject(bucketID string, objectID string) (*Object, error) {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	output, err := regionalClient.GetObject(s.ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s from bucket %s: %w", objectID, bucketID, err)
	}
	defer output.Body.Close()

	// Read the content
	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     aws.ToInt64(output.ContentLength),
		Data:     []string{string(content)},
	}, nil
}

// DeleteObject deletes an object from a bucket
func (s *AWSS3Service) DeleteObject(bucketID string, objectID string) error {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	_, err := regionalClient.DeleteObject(s.ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s from bucket %s: %w", objectID, bucketID, err)
	}

	return nil
}

// GetBucketRegion gets the region where a bucket is located
func (s *AWSS3Service) GetBucketRegion(bucketID string) (string, error) {
	output, err := s.client.GetBucketLocation(s.ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get bucket location for %s: %w", bucketID, err)
	}

	// AWS returns empty string for us-east-1
	region := string(output.LocationConstraint)
	if region == "" {
		region = "us-east-1"
	}

	return region, nil
}

// EnsureDefaultResourceExists ensures at least one S3 bucket exists for testing
// Takes the result of ListBuckets() and creates a default bucket if none exist
func (s *AWSS3Service) EnsureDefaultResourceExists(buckets []Bucket, err error) ([]Bucket, error) {
	// If there was an error listing buckets, return it
	if err != nil {
		return nil, err
	}

	// If buckets exist, return them as-is
	if len(buckets) > 0 {
		return buckets, nil
	}

	// Create a default test bucket
	defaultBucketName := fmt.Sprintf("ccc-test-bucket-%s", strings.ToLower(s.cloudParams.Region))
	fmt.Printf("📦 No buckets found. Creating default test bucket: %s\n", defaultBucketName)

	bucket, err := s.CreateBucket(defaultBucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to create default bucket: %w", err)
	}

	fmt.Printf("✅ Default bucket created successfully\n")
	return []Bucket{*bucket}, nil
}

// GetBucketRetentionDurationDays retrieves the Object Lock retention duration in days for a bucket
func (s *AWSS3Service) GetBucketRetentionDurationDays(bucketID string) (int, error) {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	// Get Object Lock configuration
	lockConfig, err := regionalClient.GetObjectLockConfiguration(s.ctx, &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		// No Object Lock configured
		return 0, nil
	}

	// Check if Object Lock is enabled with a default retention
	if lockConfig.ObjectLockConfiguration != nil &&
		lockConfig.ObjectLockConfiguration.Rule != nil &&
		lockConfig.ObjectLockConfiguration.Rule.DefaultRetention != nil {

		retention := lockConfig.ObjectLockConfiguration.Rule.DefaultRetention
		if retention.Days != nil {
			return int(*retention.Days), nil
		}
		if retention.Years != nil {
			return int(*retention.Years * 365), nil
		}
	}

	// No default retention configured
	return 0, nil
}

// GetObjectRetentionDurationDays retrieves the Object Lock retention duration in days for a specific object
func (s *AWSS3Service) GetObjectRetentionDurationDays(bucketID string, objectID string) (int, error) {
	// Create a regional client
	regionalConfig := s.config.Copy()
	regionalConfig.Region = s.cloudParams.Region
	regionalClient := s3.NewFromConfig(regionalConfig)

	// Get object retention
	retention, err := regionalClient.GetObjectRetention(s.ctx, &s3.GetObjectRetentionInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
	})
	if err != nil {
		// No retention set on this object, check bucket default
		return s.GetBucketRetentionDurationDays(bucketID)
	}

	// Calculate days until retention expires
	if retention.Retention != nil && retention.Retention.RetainUntilDate != nil {
		daysUntilExpiry := int(time.Until(*retention.Retention.RetainUntilDate).Hours() / 24)
		if daysUntilExpiry > 0 {
			return daysUntilExpiry, nil
		}
		return 0, nil // Retention already expired
	}

	// No retention set
	return 0, nil
}

// GetOrProvisionTestableResources returns all S3 buckets as testable resources
func (s *AWSS3Service) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	// List all buckets and ensure at least one exists
	buckets, err := s.EnsureDefaultResourceExists(s.ListBuckets())
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	// Convert buckets to TestParams
	resources := make([]environment.TestParams, 0, len(buckets))
	for _, bucket := range buckets {
		resources = append(resources, environment.TestParams{
			ResourceName:        bucket.Name,
			UID:                 bucket.ID,
			ProviderServiceType: "s3",
			ServiceType:         "object-storage",
			CatalogTypes:        []string{"CCC.ObjStor"},
			CloudParams:         s.cloudParams,
		})
	}

	return resources, nil
}

// CheckUserProvisioned validates that the given identity can access S3
// For AWS, credentials are immediately usable, so this just attempts a simple S3 API call
func (s *AWSS3Service) CheckUserProvisioned() error {
	// Try to list buckets as a validation that credentials work
	_, err := s.client.ListBuckets(s.ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("credentials not ready for S3 access: %w", err)
	}
	return nil
}

func (s *AWSS3Service) ElevateAccessForInspection() error {
	// No-op: AWS S3 access is managed through IAM policies, not network access
	return nil
}

// SetObjectPermission attempts to set object-level permissions using S3 ACLs
// If S3 bucket has ACLs disabled (uniform bucket-level access), this will fail
func (s *AWSS3Service) SetObjectPermission(bucketID, objectID string, permissionLevel string) error {
	// Map permission level to S3 canned ACL
	var acl string
	switch permissionLevel {
	case "read":
		acl = "public-read"
	case "write":
		acl = "public-read-write"
	case "none":
		acl = "private"
	default:
		return fmt.Errorf("unsupported permission level: %s", permissionLevel)
	}

	// Attempt to set object-level ACL
	// If bucket has ACLs disabled (enforcing uniform access), this will fail
	_, err := s.client.PutObjectAcl(s.ctx, &s3.PutObjectAclInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
		ACL:    s3types.ObjectCannedACL(acl),
	})

	if err != nil {
		// Check if it's because ACLs are disabled (which is GOOD - uniform access is enforced)
		if strings.Contains(err.Error(), "AccessControlListNotSupported") {
			return fmt.Errorf("S3 object-level ACLs are disabled - uniform bucket-level access is enforced: %w", err)
		}
		return fmt.Errorf("failed to set object ACL: %w", err)
	}

	// ACL was set successfully (only happens if uniform access is NOT enforced)
	return nil
}

// ListDeletedBuckets returns an error - AWS S3 does not support bucket-level soft delete
// S3 bucket deletion is immediate and permanent (CN03.AR01 not supported)
func (s *AWSS3Service) ListDeletedBuckets() ([]Bucket, error) {
	return nil, fmt.Errorf("AWS S3 does not support bucket-level soft delete - bucket deletion is immediate and permanent")
}

// RestoreBucket returns an error - AWS S3 does not support bucket-level soft delete
// S3 bucket deletion is immediate and permanent (CN03.AR01 not supported)
func (s *AWSS3Service) RestoreBucket(bucketID string) error {
	return fmt.Errorf("AWS S3 does not support bucket restoration - bucket deletion is immediate and permanent")
}

// SetBucketRetentionDurationDays returns an error - AWS S3 does not support bucket-level retention policies
// S3 has Object Lock for object-level retention, but not bucket-level (CN03.AR02 not supported at bucket level)
func (s *AWSS3Service) SetBucketRetentionDurationDays(bucketID string, days int) error {
	return fmt.Errorf("AWS S3 does not support bucket-level retention policies - use Object Lock for object-level retention instead")
}

// ResetAccess is a no-op for AWS S3 (access is managed via IAM)
func (s *AWSS3Service) ResetAccess() error {
	// No-op: AWS S3 access is managed through IAM policies, not network access
	return nil
}
