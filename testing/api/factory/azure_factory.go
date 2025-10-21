package factory

import (
	"context"
	"fmt"
	"os"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

// AzureFactory implements the Factory interface for Azure
type AzureFactory struct {
	ctx context.Context
}

// NewAzureFactory creates a new Azure factory
func NewAzureFactory() *AzureFactory {
	return &AzureFactory{
		ctx: context.Background(),
	}
}

// GetServiceAPI returns a service API client for the given service type (string version for Gherkin)
func (f *AzureFactory) GetServiceAPI(serviceType string) (any, error) {
	return f.GetServiceAPIForType(ServiceType(serviceType))
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity (string version for Gherkin)
func (f *AzureFactory) GetServiceAPIWithIdentity(serviceType string, identity *iam.Identity) (any, error) {
	return f.GetServiceAPIWithIdentityForType(ServiceType(serviceType), identity)
}

// GetServiceAPIForType returns a service API client for the given service type (typed version for Go code)
func (f *AzureFactory) GetServiceAPIForType(serviceType ServiceType) (any, error) {
	// Get subscription ID and resource group from environment
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	resourceGroup := os.Getenv("AZURE_RESOURCE_GROUP")

	switch serviceType {
	case ServiceTypeIAM:
		return iam.NewAzureIAMService(f.ctx, subscriptionID, resourceGroup)
	case ServiceTypeObjectStorage:
		// TODO: Implement Azure Blob Storage service creation
		return nil, fmt.Errorf("object-storage not yet implemented for Azure")
	default:
		return nil, fmt.Errorf("unsupported service type for Azure: %s", serviceType)
	}
}

// GetServiceAPIWithIdentityForType returns a service API client authenticated as the given identity (typed version for Go code)
func (f *AzureFactory) GetServiceAPIWithIdentityForType(serviceType ServiceType, identity *iam.Identity) (any, error) {
	if identity.Provider != string(ProviderAzure) {
		return nil, fmt.Errorf("identity is not for Azure provider: %s", identity.Provider)
	}

	// Get subscription ID and resource group from identity or environment
	subscriptionID := identity.Credentials["subscription_id"]
	resourceGroup := identity.Credentials["resource_group"]
	if subscriptionID == "" {
		subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	}
	if resourceGroup == "" {
		resourceGroup = os.Getenv("AZURE_RESOURCE_GROUP")
	}

	switch serviceType {
	case ServiceTypeIAM:
		// IAM service doesn't typically use per-identity clients, return the standard IAM service
		return iam.NewAzureIAMService(f.ctx, subscriptionID, resourceGroup)

	case ServiceTypeObjectStorage:
		// TODO: Implement Azure Blob Storage service with credentials
		return nil, fmt.Errorf("object-storage with identity not yet implemented for Azure")

	default:
		return nil, fmt.Errorf("unsupported service type for Azure: %s", serviceType)
	}
}

// GetProvider returns the cloud provider
func (f *AzureFactory) GetProvider() CloudProvider {
	return ProviderAzure
}

// SetContext sets the context for this factory
func (f *AzureFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}
