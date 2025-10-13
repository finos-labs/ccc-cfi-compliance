package factory

import (
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
)

// AzureFactory implements the Factory interface for Azure
type AzureFactory struct {
	subscriptionID string
	config         map[string]interface{}
}

// NewAzureFactory creates a new Azure factory
func NewAzureFactory() *AzureFactory {
	return &AzureFactory{
		config: make(map[string]interface{}),
	}
}

// GetServiceAPI returns a generic service API client for the given Azure resource ID
func (f *AzureFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	// TODO: Implement Azure service client creation
	return nil, fmt.Errorf("GetServiceAPI not yet implemented for Azure: %s", serviceID)
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AzureFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	// TODO: Implement Azure service client creation with specific identity
	if identity.Provider != string(ProviderAzure) {
		return nil, fmt.Errorf("identity is not for Azure provider: %s", identity.Provider)
	}
	return nil, fmt.Errorf("GetServiceAPIWithIdentity not yet implemented for Azure: %s", serviceID)
}

// GetIAMService returns the IAM service for Azure
func (f *AzureFactory) GetIAMService() (iam.IAMService, error) {
	// TODO: Implement Azure IAM service
	return nil, fmt.Errorf("GetIAMService not yet implemented for Azure")
}

// GetObjectStorageService returns an Azure Blob Storage service for the given service ID
func (f *AzureFactory) GetObjectStorageService(serviceID string) (objstorage.Service, error) {
	// TODO: Implement Azure Blob Storage service creation
	return nil, fmt.Errorf("GetObjectStorageService not yet implemented for Azure: %s", serviceID)
}

// GetObjectStorageServiceWithIdentity returns an Azure Blob Storage service authenticated as the given identity
func (f *AzureFactory) GetObjectStorageServiceWithIdentity(serviceID string, identity *iam.Identity) (objstorage.Service, error) {
	// TODO: Implement Azure Blob Storage service creation with specific identity
	if identity.Provider != string(ProviderAzure) {
		return nil, fmt.Errorf("identity is not for Azure provider: %s", identity.Provider)
	}
	return nil, fmt.Errorf("GetObjectStorageServiceWithIdentity not yet implemented for Azure: %s", serviceID)
}

// GetProvider returns the cloud provider
func (f *AzureFactory) GetProvider() CloudProvider {
	return ProviderAzure
}

// SetSubscriptionID sets the Azure subscription ID for this factory
func (f *AzureFactory) SetSubscriptionID(subscriptionID string) {
	f.subscriptionID = subscriptionID
}
