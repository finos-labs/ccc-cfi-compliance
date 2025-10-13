package factory

import (
	"context"
	"fmt"
	"os"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
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

// GetServiceAPI returns a generic service API client for the given service type
func (f *AzureFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	var service generic.Service
	var err error

	// Get subscription ID and resource group from environment
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	resourceGroup := os.Getenv("AZURE_RESOURCE_GROUP")

	switch serviceID {
	case "iam":
		service, err = iam.NewAzureIAMService(f.ctx, subscriptionID, resourceGroup)
	case "object-storage":
		// TODO: Implement Azure Blob Storage service creation
		return nil, fmt.Errorf("object-storage not yet implemented for Azure")
	default:
		return nil, fmt.Errorf("unsupported service type for Azure: %s", serviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Azure service '%s': %w", serviceID, err)
	}

	return service, nil
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AzureFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	if identity.Provider != string(ProviderAzure) {
		return nil, fmt.Errorf("identity is not for Azure provider: %s", identity.Provider)
	}

	var service generic.Service
	var err error

	// Get subscription ID and resource group from identity or environment
	subscriptionID := identity.Credentials["subscription_id"]
	resourceGroup := identity.Credentials["resource_group"]
	if subscriptionID == "" {
		subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	}
	if resourceGroup == "" {
		resourceGroup = os.Getenv("AZURE_RESOURCE_GROUP")
	}

	switch serviceID {
	case "iam":
		// IAM service doesn't typically use per-identity clients, return the standard IAM service
		service, err = iam.NewAzureIAMService(f.ctx, subscriptionID, resourceGroup)

	case "object-storage":
		// TODO: Implement Azure Blob Storage service with credentials
		return nil, fmt.Errorf("object-storage with identity not yet implemented for Azure")

	default:
		return nil, fmt.Errorf("unsupported service type for Azure: %s", serviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Azure service '%s' with identity: %w", serviceID, err)
	}

	return service, nil
}

// GetProvider returns the cloud provider
func (f *AzureFactory) GetProvider() CloudProvider {
	return ProviderAzure
}

// SetContext sets the context for this factory
func (f *AzureFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}
