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
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AzureBlobService implements Service for Azure Blob Storage
type AzureBlobService struct {
	storageClient *armstorage.AccountsClient
	credential    azcore.TokenCredential
	ctx           context.Context
	cloudParams   environment.CloudParams
}

// NewAzureBlobService creates a new Azure Blob Storage service using default credentials
func NewAzureBlobService(ctx context.Context, cloudParams environment.CloudParams) (*AzureBlobService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	storageClient, err := armstorage.NewAccountsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	return &AzureBlobService{
		storageClient: storageClient,
		credential:    cred,
		ctx:           ctx,
		cloudParams:   cloudParams,
	}, nil
}

// NewAzureBlobServiceWithCredentials creates a new Azure Blob Storage service with specific credentials
func NewAzureBlobServiceWithCredentials(ctx context.Context, cloudParams environment.CloudParams, identity *iam.Identity) (*AzureBlobService, error) {
	// For managed identities, we need to create a credential using the client ID
	clientID := identity.Credentials["client_id"]
	if clientID == "" {
		return nil, fmt.Errorf("client_id not found in identity credentials")
	}

	// Create managed identity credential
	cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(clientID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create managed identity credential: %w", err)
	}

	storageClient, err := armstorage.NewAccountsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	fmt.Printf("🔐 Created Azure Blob Storage client with managed identity:\n")
	fmt.Printf("   Client ID: %s\n", clientID)

	return &AzureBlobService{
		storageClient: storageClient,
		credential:    cred,
		ctx:           ctx,
		cloudParams:   cloudParams,
	}, nil
}

// ListBuckets lists all storage accounts and their containers in the configured resource group
// In Azure, a "bucket" is represented as "resourceGroup/storageAccount/containerName"
func (s *AzureBlobService) ListBuckets() ([]Bucket, error) {
	buckets := []Bucket{}

	// List all storage accounts in the resource group
	pager := s.storageClient.NewListByResourceGroupPager(s.cloudParams.AzureResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list storage accounts: %w", err)
		}

		for _, account := range page.Value {
			if account.Name == nil || account.Location == nil {
				continue
			}

			accountName := *account.Name
			location := *account.Location
			resourceGroup := s.cloudParams.AzureResourceGroup

			// For each storage account, list its containers
			containers, err := s.listContainersForAccount(accountName)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to list containers for %s: %v\n", accountName, err)
				continue
			}

			// If no containers, still add the storage account as a bucket
			if len(containers) == 0 {
				buckets = append(buckets, Bucket{
					ID:     fmt.Sprintf("%s/%s", resourceGroup, accountName),
					Name:   accountName,
					Region: location,
				})
			} else {
				// Add each container as a separate bucket
				for _, containerName := range containers {
					bucketID := fmt.Sprintf("%s/%s/%s", resourceGroup, accountName, containerName)
					buckets = append(buckets, Bucket{
						ID:     bucketID,
						Name:   containerName,
						Region: location,
					})
				}
			}
		}
	}

	return buckets, nil
}

// CreateBucket creates a new storage account and container
// bucketID format: "resourceGroup/storageAccountName" or "resourceGroup/storageAccountName/containerName"
func (s *AzureBlobService) CreateBucket(bucketID string) (*Bucket, error) {
	// Parse bucketID to extract resource group, storage account and container names
	parts := strings.Split(bucketID, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("bucketID must include resource group: resourceGroup/storageAccount or resourceGroup/storageAccount/container")
	}

	resourceGroup := parts[0]
	storageAccountName := parts[1]
	containerName := "default"
	if len(parts) > 2 {
		containerName = parts[2]
	}

	// Sanitize storage account name (lowercase, alphanumeric, 3-24 chars)
	storageAccountName = sanitizeStorageAccountName(storageAccountName)

	fmt.Printf("📦 Creating storage account %s in resource group %s with container %s in %s...\n", storageAccountName, resourceGroup, containerName, s.cloudParams.Region)

	// Create storage account
	sku := armstorage.SKUNameStandardLRS
	kind := armstorage.KindStorageV2
	allowBlobPublicAccess := false

	location := s.cloudParams.Region
	params := armstorage.AccountCreateParameters{
		Location: &location,
		SKU: &armstorage.SKU{
			Name: &sku,
		},
		Kind: &kind,
		Properties: &armstorage.AccountPropertiesCreateParameters{
			AllowBlobPublicAccess: &allowBlobPublicAccess,
			AccessTier:            toAccessTierPtr(armstorage.AccessTierHot),
		},
		Tags: map[string]*string{
			"Purpose":   toStringPtr("CCC-Testing"),
			"ManagedBy": toStringPtr("CCC-CFI-Compliance-Framework"),
		},
	}

	poller, err := s.storageClient.BeginCreate(s.ctx, resourceGroup, storageAccountName, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin creating storage account %s: %w", storageAccountName, err)
	}

	_, err = poller.PollUntilDone(s.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage account %s: %w", storageAccountName, err)
	}

	fmt.Printf("   ✅ Storage account created\n")

	// Create container
	err = s.createContainer(storageAccountName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	fmt.Printf("   ✅ Container created\n")

	return &Bucket{
		ID:     fmt.Sprintf("%s/%s/%s", resourceGroup, storageAccountName, containerName),
		Name:   containerName,
		Region: s.cloudParams.Region,
	}, nil
}

// DeleteBucket deletes a storage account or container
// bucketID format: "resourceGroup/storageAccount" or "resourceGroup/storageAccount/container"
func (s *AzureBlobService) DeleteBucket(bucketID string) error {
	parts := strings.Split(bucketID, "/")
	if len(parts) < 2 {
		return fmt.Errorf("bucketID must include resource group: resourceGroup/storageAccount or resourceGroup/storageAccount/container")
	}

	resourceGroup := parts[0]
	storageAccountName := parts[1]

	if len(parts) > 2 {
		// Delete specific container
		containerName := parts[2]
		fmt.Printf("🗑️  Deleting container %s from storage account %s...\n", containerName, storageAccountName)
		return s.deleteContainer(storageAccountName, containerName)
	}

	// Delete entire storage account
	fmt.Printf("🗑️  Deleting storage account %s from resource group %s...\n", storageAccountName, resourceGroup)
	_, err := s.storageClient.Delete(s.ctx, resourceGroup, storageAccountName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete storage account %s: %w", storageAccountName, err)
	}

	return nil
}

// GetBucketRegion returns the region where a bucket (storage account) is located
// bucketID format: "resourceGroup/storageAccount" or "resourceGroup/storageAccount/container"
func (s *AzureBlobService) GetBucketRegion(bucketID string) (string, error) {
	parts := strings.Split(bucketID, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("bucketID must include resource group: resourceGroup/storageAccount")
	}

	resourceGroup := parts[0]
	storageAccountName := parts[1]

	account, err := s.storageClient.GetProperties(s.ctx, resourceGroup, storageAccountName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get storage account properties: %w", err)
	}

	if account.Location == nil {
		return "", fmt.Errorf("storage account location is nil")
	}

	return *account.Location, nil
}

// ListObjects lists all blobs in a container
// bucketID format: "resourceGroup/storageAccount/container"
func (s *AzureBlobService) ListObjects(bucketID string) ([]Object, error) {
	parts := strings.Split(bucketID, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("bucketID must include resource group and container: resourceGroup/storageAccount/container")
	}

	storageAccountName := parts[1]
	containerName := parts[2]

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
// bucketID format: "resourceGroup/storageAccount/container"
func (s *AzureBlobService) CreateObject(bucketID string, objectID string, data []string) (*Object, error) {
	parts := strings.Split(bucketID, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("bucketID must include resource group and container: resourceGroup/storageAccount/container")
	}

	storageAccountName := parts[1]
	containerName := parts[2]

	// Convert []string to []byte
	var content bytes.Buffer
	for _, line := range data {
		content.WriteString(line)
	}

	// Get blob client
	blobClient, err := s.getBlobServiceClient(storageAccountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob service client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(objectID)

	// Upload blob
	_, err = blockBlobClient.UploadStream(s.ctx, bytes.NewReader(content.Bytes()), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob %s: %w", objectID, err)
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     int64(content.Len()),
		Data:     data,
	}, nil
}

// ReadObject reads a blob from a container
// bucketID format: "resourceGroup/storageAccount/container"
func (s *AzureBlobService) ReadObject(bucketID string, objectID string) (*Object, error) {
	parts := strings.Split(bucketID, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("bucketID must include resource group and container: resourceGroup/storageAccount/container")
	}

	storageAccountName := parts[1]
	containerName := parts[2]

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
// bucketID format: "resourceGroup/storageAccount/container"
func (s *AzureBlobService) DeleteObject(bucketID string, objectID string) error {
	parts := strings.Split(bucketID, "/")
	if len(parts) < 3 {
		return fmt.Errorf("bucketID must include resource group and container: resourceGroup/storageAccount/container")
	}

	storageAccountName := parts[1]
	containerName := parts[2]

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

// sanitizeStorageAccountName ensures the name meets Azure requirements
// Storage account names must be 3-24 characters, lowercase letters and numbers only
func sanitizeStorageAccountName(name string) string {
	result := ""
	for _, char := range strings.ToLower(name) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result += string(char)
		}
	}

	// Ensure minimum length
	if len(result) < 3 {
		result = result + "test"
	}

	// Ensure maximum length
	if len(result) > 24 {
		result = result[:24]
	}

	return result
}

// extractResourceGroupFromID extracts the resource group name from an Azure resource ID
// Resource ID format: /subscriptions/{sub}/resourceGroups/{rg}/providers/...
func extractResourceGroupFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// Helper pointer functions
func toStringPtr(s string) *string {
	return &s
}

func toAccessTierPtr(tier armstorage.AccessTier) *armstorage.AccessTier {
	return &tier
}

// GetTestableResources returns all Azure storage containers as testable resources
func (s *AzureBlobService) GetTestableResources() ([]environment.TestParams, error) {
	// List all buckets (storage accounts + containers)
	buckets, err := s.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Convert containers to TestParams
	resources := make([]environment.TestParams, 0, len(buckets))
	for _, bucket := range buckets {
		resources = append(resources, environment.TestParams{
			ResourceName:        bucket.Name,
			UID:                 bucket.ID,
			ProviderServiceType: "Microsoft.Storage/storageAccounts",
			CatalogTypes:        []string{"CCC.ObjStor", "CCC.Core"},
			CloudParams:         s.cloudParams,
		})
	}

	return resources, nil
}
