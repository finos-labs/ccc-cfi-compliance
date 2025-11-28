package factory

import (
	"context"
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AzureFactory implements the Factory interface for Azure
type AzureFactory struct {
	ctx         context.Context
	cloudParams environment.CloudParams
}

// NewAzureFactory creates a new Azure factory
func NewAzureFactory(cloudParams environment.CloudParams) *AzureFactory {
	return &AzureFactory{
		ctx:         context.Background(),
		cloudParams: cloudParams,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *AzureFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	var service generic.Service
	var err error

	switch serviceID {
	case "iam":
		service, err = iam.NewAzureIAMService(f.ctx, f.cloudParams)
	case "object-storage":
		service, err = objstorage.NewAzureBlobService(f.ctx, f.cloudParams)
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
	switch serviceID {
	case "iam":
		// IAM service doesn't typically use per-identity clients, return the standard IAM service
		service, err = iam.NewAzureIAMService(f.ctx, f.cloudParams)

	case "object-storage":
		service, err = objstorage.NewAzureBlobServiceWithCredentials(f.ctx, f.cloudParams, identity)

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
