package factory

import (
	"context"
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
)

// AWSFactory implements the Factory interface for AWS
type AWSFactory struct {
	ctx context.Context
}

// NewAWSFactory creates a new AWS factory
func NewAWSFactory() *AWSFactory {
	return &AWSFactory{
		ctx: context.Background(),
	}
}

// GetServiceAPI returns a service API client for the given service type (string version for Gherkin)
func (f *AWSFactory) GetServiceAPI(serviceType string) (any, error) {
	return f.GetServiceAPIForType(ServiceType(serviceType))
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity (string version for Gherkin)
func (f *AWSFactory) GetServiceAPIWithIdentity(serviceType string, identity *iam.Identity) (any, error) {
	return f.GetServiceAPIWithIdentityForType(ServiceType(serviceType), identity)
}

// GetServiceAPIForType returns a service API client for the given service type (typed version for Go code)
func (f *AWSFactory) GetServiceAPIForType(serviceType ServiceType) (any, error) {
	switch serviceType {
	case ServiceTypeIAM:
		return iam.NewAWSIAMService(f.ctx)
	case ServiceTypeObjectStorage:
		return objstorage.NewAWSS3Service(f.ctx)
	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceType)
	}
}

// GetServiceAPIWithIdentityForType returns a service API client authenticated as the given identity (typed version for Go code)
func (f *AWSFactory) GetServiceAPIWithIdentityForType(serviceType ServiceType, identity *iam.Identity) (any, error) {
	if identity.Provider != string(ProviderAWS) {
		return nil, fmt.Errorf("identity is not for AWS provider: %s", identity.Provider)
	}

	switch serviceType {
	case ServiceTypeIAM:
		// IAM service doesn't typically use per-identity clients, return the standard IAM service
		return iam.NewAWSIAMService(f.ctx)

	case ServiceTypeObjectStorage:
		return objstorage.NewAWSS3ServiceWithCredentials(f.ctx, identity)

	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceType)
	}
}

// GetProvider returns the cloud provider
func (f *AWSFactory) GetProvider() CloudProvider {
	return ProviderAWS
}

// SetContext sets the context for this factory
func (f *AWSFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}
