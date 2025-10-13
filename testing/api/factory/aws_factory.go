package factory

import (
	"context"
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
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

// GetServiceAPI returns a generic service API client for the given service type
func (f *AWSFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	var service generic.Service
	var err error

	switch serviceID {
	case "iam":
		service, err = iam.NewAWSIAMService(f.ctx)
	case "object-storage":
		service, err = objstorage.NewAWSS3Service(f.ctx)
	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS service '%s': %w", serviceID, err)
	}

	return service, nil
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AWSFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	if identity.Provider != string(ProviderAWS) {
		return nil, fmt.Errorf("identity is not for AWS provider: %s", identity.Provider)
	}

	var service generic.Service
	var err error

	switch serviceID {
	case "iam":
		// IAM service doesn't typically use per-identity clients, return the standard IAM service
		service, err = iam.NewAWSIAMService(f.ctx)

	case "object-storage":
		service, err = objstorage.NewAWSS3ServiceWithCredentials(f.ctx, identity)

	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS service '%s' with identity: %w", serviceID, err)
	}

	return service, nil
}

// GetProvider returns the cloud provider
func (f *AWSFactory) GetProvider() CloudProvider {
	return ProviderAWS
}

// SetContext sets the context for this factory
func (f *AWSFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}
