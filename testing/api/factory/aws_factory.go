package factory

import (
	"context"
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AWSFactory implements the Factory interface for AWS
type AWSFactory struct {
	ctx         context.Context
	cloudParams environment.CloudParams
	iamService  generic.Service
}

// NewAWSFactory creates a new AWS factory
func NewAWSFactory(cloudParams environment.CloudParams) *AWSFactory {
	ctx := context.Background()

	// Create IAM service once and cache it
	iamService, err := iam.NewAWSIAMService(ctx)
	if err != nil {
		// Log error but don't fail - IAM service might not be needed
		fmt.Printf("⚠️  Warning: Failed to create AWS IAM service: %v\n", err)
	}

	return &AWSFactory{
		ctx:         ctx,
		cloudParams: cloudParams,
		iamService:  iamService,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *AWSFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	switch serviceID {
	case "iam":
		if f.iamService == nil {
			return nil, fmt.Errorf("AWS IAM service not initialized")
		}
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewAWSS3Service(f.ctx, f.cloudParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS service '%s': %w", serviceID, err)
		}

		// Elevate access for testing
		if err := service.ElevateAccessForInspection(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to elevate access for %s: %v\n", serviceID, err)
		}

		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceID)
	}
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AWSFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	if identity.Provider != string(ProviderAWS) {
		return nil, fmt.Errorf("identity is not for AWS provider: %s", identity.Provider)
	}

	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewAWSS3ServiceWithCredentials(f.ctx, f.cloudParams, identity)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS service '%s' with identity: %w", serviceID, err)
		}

		// Elevate access for testing
		if err := service.ElevateAccessForInspection(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to elevate access for %s: %v\n", serviceID, err)
		}

		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceID)
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
