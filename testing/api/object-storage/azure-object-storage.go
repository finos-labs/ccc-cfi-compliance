package objstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage/elevation"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AzureBlobService implements Service for Azure Blob Storage
type AzureBlobService struct {
	storageClient *armstorage.AccountsClient // For normal storage operations
	credential    azcore.TokenCredential
	ctx           context.Context
	cloudParams   environment.CloudParams
	elevator      *elevation.AzureStorageElevator // Handles access elevation (RBAC + network)
}

// NewAzureBlobService creates a new Azure Blob Storage service using default credentials
func NewAzureBlobService(ctx context.Context, cloudParams environment.CloudParams) (*AzureBlobService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Create storage client for normal operations
	storageClient, err := armstorage.NewAccountsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	// Create elevator for managing access controls (RBAC + network)
	elevator, err := elevation.NewAzureStorageElevator(
		ctx,
		cred,
		cloudParams.AzureSubscriptionID,
		cloudParams.AzureResourceGroup,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure storage elevator: %w", err)
	}

	return &AzureBlobService{
		storageClient: storageClient,
		credential:    cred,
		ctx:           ctx,
		cloudParams:   cloudParams,
		elevator:      elevator,
	}, nil
}

// NewAzureBlobServiceWithCredentials creates a new Azure Blob Storage service with service principal credentials
func NewAzureBlobServiceWithCredentials(ctx context.Context, cloudParams environment.CloudParams, identity *iam.Identity) (*AzureBlobService, error) {
	// Extract service principal credentials
	clientID := identity.Credentials["client_id"]
	if clientID == "" {
		return nil, fmt.Errorf("client_id not found in identity credentials")
	}

	clientSecret := identity.Credentials["client_secret"]
	if clientSecret == "" {
		return nil, fmt.Errorf("client_secret not found in identity credentials")
	}

	tenantID := identity.Credentials["tenant_id"]
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id not found in identity credentials")
	}

	fmt.Printf("🔐 Creating Azure Blob Storage client with service principal:\n")
	fmt.Printf("   Client ID: %s\n", clientID)
	fmt.Printf("   Tenant ID: %s\n", tenantID)

	// Create service principal credential
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create service principal credential: %w", err)
	}

	// Create storage client for normal operations
	storageClient, err := armstorage.NewAccountsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	// Create elevator for managing access controls (RBAC + network)
	elevator, err := elevation.NewAzureStorageElevator(
		ctx,
		cred,
		cloudParams.AzureSubscriptionID,
		cloudParams.AzureResourceGroup,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure storage elevator: %w", err)
	}

	return &AzureBlobService{
		storageClient: storageClient,
		credential:    cred,
		ctx:           ctx,
		cloudParams:   cloudParams,
		elevator:      elevator,
	}, nil
}

// ListBuckets lists all containers in the identified storage account
// In Azure, a "bucket" is represented as "resourceGroup/storageAccount/containerName"
func (s *AzureBlobService) ListBuckets() ([]Bucket, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	fmt.Printf("📦 Using storage account: %s\n", storageAccountName)

	buckets := []Bucket{}
	resourceGroup := s.cloudParams.AzureResourceGroup

	// Get the storage account location
	account, err := s.storageClient.GetProperties(s.ctx, resourceGroup, storageAccountName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage account properties: %w", err)
	}

	location := s.cloudParams.Region
	if account.Location != nil {
		location = *account.Location
	}

	// List containers in the storage account
	containers, err := s.listContainersForAccount(storageAccountName)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for %s: %w", storageAccountName, err)
	}

	// Add each container as a separate bucket (ID is just the container name)
	for _, containerName := range containers {
		buckets = append(buckets, Bucket{
			ID:     containerName,
			Name:   containerName,
			Region: location,
		})
	}

	return buckets, nil
}

// CreateBucket creates a new container in the storage account
// bucketID is the container name
func (s *AzureBlobService) CreateBucket(bucketID string) (*Bucket, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID
	fmt.Printf("📦 Creating container %s in storage account %s...\n", containerName, storageAccountName)

	// Create container in the existing storage account
	err := s.createContainer(storageAccountName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	fmt.Printf("   ✅ Container created\n")

	return &Bucket{
		ID:     containerName,
		Name:   containerName,
		Region: s.cloudParams.Region,
	}, nil
}

// DeleteBucket deletes a container from the storage account
// bucketID is the container name
func (s *AzureBlobService) DeleteBucket(bucketID string) error {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID
	fmt.Printf("🗑️  Deleting container %s from storage account %s...\n", containerName, storageAccountName)
	return s.deleteContainer(storageAccountName, containerName)
}

// GetBucketRegion returns the region where the storage account is located
func (s *AzureBlobService) GetBucketRegion(bucketID string) (string, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	account, err := s.storageClient.GetProperties(s.ctx, s.cloudParams.AzureResourceGroup, storageAccountName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get storage account properties: %w", err)
	}

	if account.Location == nil {
		return "", fmt.Errorf("storage account location is nil")
	}

	return *account.Location, nil
}

// ListObjects lists all blobs in a container
// bucketID is the container name
func (s *AzureBlobService) ListObjects(bucketID string) ([]Object, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID

	// Get blob service client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	pager := containerClient.NewListBlobsFlatPager(nil)

	objects := []Object{}
	for pager.More() {
		page, err := pager.NextPage(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			if blob.Name == nil {
				continue
			}

			size := int64(0)
			if blob.Properties != nil && blob.Properties.ContentLength != nil {
				size = *blob.Properties.ContentLength
			}

			objects = append(objects, Object{
				ID:       *blob.Name,
				BucketID: bucketID,
				Name:     *blob.Name,
				Size:     size,
				Data:     nil,
			})
		}
	}

	return objects, nil
}

// CreateObject creates a new blob in a container
// bucketID is the container name
func (s *AzureBlobService) CreateObject(bucketID string, objectID string, data string) (*Object, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID

	// Convert string to []byte
	content := []byte(data)

	// Get blob client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(objectID)

	// Upload blob
	_, err = blockBlobClient.UploadStream(s.ctx, bytes.NewReader(content), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob %s: %w", objectID, err)
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     int64(len(content)),
		Data:     []string{data},
	}, nil
}

// ReadObject reads a blob from a container
// bucketID is the container name
func (s *AzureBlobService) ReadObject(bucketID string, objectID string) (*Object, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID

	// Get blob client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(objectID)

	// Download blob
	downloadResponse, err := blockBlobClient.DownloadStream(s.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob %s: %w", objectID, err)
	}
	defer downloadResponse.Body.Close()

	// Read content
	content, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob content: %w", err)
	}

	size := int64(len(content))
	if downloadResponse.ContentLength != nil {
		size = *downloadResponse.ContentLength
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     size,
		Data:     []string{string(content)},
	}, nil
}

// DeleteObject deletes a blob from a container
// bucketID is the container name
func (s *AzureBlobService) DeleteObject(bucketID string, objectID string) error {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID

	// Get blob client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(objectID)

	// Delete blob
	_, err = blockBlobClient.Delete(s.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob %s: %w", objectID, err)
	}

	return nil
}

// Helper functions

// getBlobServiceClient creates a blob service client for a storage account
func (s *AzureBlobService) getBlobServiceClient(storageAccountName string) (*azblob.Client, error) {
	// Construct the blob service URL
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName)

	// Create blob client with Azure AD authentication
	client, err := azblob.NewClient(serviceURL, s.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client: %w", err)
	}

	return client, nil
}

// listContainersForAccount lists all containers in a storage account
func (s *AzureBlobService) listContainersForAccount(storageAccountName string) ([]string, error) {
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return nil, err
	}

	containers := []string{}
	pager := blobClient.NewListContainersPager(nil)

	for pager.More() {
		page, err := pager.NextPage(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list containers: %w", err)
		}

		for _, cont := range page.ContainerItems {
			if cont.Name != nil {
				containers = append(containers, *cont.Name)
			}
		}
	}

	return containers, nil
}

// createContainer creates a new container in a storage account
func (s *AzureBlobService) createContainer(storageAccountName, containerName string) error {
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return err
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	_, err = containerClient.Create(s.ctx, &container.CreateOptions{})
	if err != nil {
		// Check if container already exists
		if strings.Contains(err.Error(), "ContainerAlreadyExists") {
			fmt.Printf("   ℹ️  Container already exists\n")
			return nil
		}
		return fmt.Errorf("failed to create container %s: %w", containerName, err)
	}

	return nil
}

// deleteContainer deletes a container from a storage account
func (s *AzureBlobService) deleteContainer(storageAccountName, containerName string) error {
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return err
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	_, err = containerClient.Delete(s.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete container %s: %w", containerName, err)
	}

	return nil
}

// EnsureDefaultResourceExists ensures at least one container exists in each storage account for testing
// Takes the result of ListBuckets() and creates default containers if needed
func (s *AzureBlobService) EnsureDefaultResourceExists(buckets []Bucket, err error) ([]Bucket, error) {
	// If there was an error listing buckets, return it
	if err != nil {
		return nil, err
	}

	// If we have any buckets/containers, return them as-is
	if len(buckets) > 0 {
		return buckets, nil
	}

	// No containers found - create a default container in the identified storage account
	fmt.Printf("📦 No containers found. Creating default container...\n")

	defaultContainerName := "ccc-test-container"
	fmt.Printf("   Creating container: %s\n", defaultContainerName)

	bucket, err := s.CreateBucket(defaultContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create default container: %w", err)
	}

	newBuckets := []Bucket{*bucket}

	fmt.Printf("✅ Default containers created successfully\n")
	return newBuckets, nil
}

// GetBucketRetentionDurationDays retrieves the retention policy duration in days for a container
func (s *AzureBlobService) GetBucketRetentionDurationDays(bucketID string) (int, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID

	// Get container client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return 0, fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)

	// Get container properties
	_, err = containerClient.GetProperties(s.ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get container properties: %w", err)
	}

	// Note: Checking immutability policies requires the Azure Blob Storage Management API
	// For now, return 0 (no retention) as the default
	// In production, this would query the container's immutability policy settings
	// which are configured via ARM templates or the Management API
	return 0, nil
}

// GetObjectRetentionDurationDays retrieves the retention policy duration in days for a blob
func (s *AzureBlobService) GetObjectRetentionDurationDays(bucketID string, objectID string) (int, error) {
	storageAccountName := s.cloudParams.AzureStorageAccount
	containerName := bucketID

	// Get blob client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return 0, fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(objectID)

	// Get blob properties
	props, err := blockBlobClient.GetProperties(s.ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get blob properties: %w", err)
	}

	// Check for legal hold or immutability policy
	if props.LegalHold != nil && *props.LegalHold {
		// Legal hold is active - return max retention
		return 9999, nil
	}

	// Check for time-based retention
	// Azure blob immutability policies inherit from container level
	// For object-level specifics, we'd need to check the blob's immutability policy
	// Return container-level retention as default
	return s.GetBucketRetentionDurationDays(bucketID)
}

// GetOrProvisionTestableResources returns all Azure storage containers as testable resources
func (s *AzureBlobService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	// Validate that storage account name is set
	if s.cloudParams.AzureStorageAccount == "" {
		return nil, fmt.Errorf("AzureStorageAccount not set in CloudParams")
	}

	// Build the storage account resource ID for RBAC
	storageAccountResourceID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s",
		s.cloudParams.AzureSubscriptionID,
		s.cloudParams.AzureResourceGroup,
		s.cloudParams.AzureStorageAccount)

	fmt.Printf("   Storage Account Resource ID for RBAC: %s\n", storageAccountResourceID)

	// List all buckets and ensure at least one container exists per storage account
	buckets, err := s.EnsureDefaultResourceExists(s.ListBuckets())
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Convert containers to TestParams
	// UID is the storage account resource ID (for RBAC scope)
	// ResourceName is the container name (for test identification)
	resources := make([]environment.TestParams, 0, len(buckets))
	for _, bucket := range buckets {
		resources = append(resources, environment.TestParams{
			ResourceName:        bucket.Name,
			UID:                 storageAccountResourceID, // Use storage account resource ID for RBAC
			ServiceType:         "object-storage",
			ProviderServiceType: "Microsoft.Storage/storageAccounts",
			CatalogTypes:        []string{"CCC.ObjStor"},
			CloudParams:         s.cloudParams,
		})
	}

	return resources, nil
}

// CheckUserProvisioned validates that the given identity can access Azure Blob Storage
// This performs a simple list operation to ensure credentials have propagated
func (s *AzureBlobService) CheckUserProvisioned() error {
	_, err := s.listContainersForAccount(s.cloudParams.AzureStorageAccount)
	if err != nil {
		return fmt.Errorf("credentials not ready for Azure Blob Storage access: %w", err)
	}
	return nil
}

func (s *AzureBlobService) ElevateAccessForInspection() error {
	return s.elevator.ElevateStorageAccountAccess(s.cloudParams.AzureStorageAccount)
}

func (s *AzureBlobService) ResetAccess() error {
	return s.elevator.ResetStorageAccountAccess(s.cloudParams.AzureStorageAccount)
}
