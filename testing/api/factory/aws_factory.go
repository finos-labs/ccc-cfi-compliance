package factory

import (
	"context"
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/logging"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
	"github.com/finos-labs/ccc-cfi-compliance/testing/types"
)

// AWSFactory implements the Factory interface for AWS
type AWSFactory struct {
	ctx        context.Context
	instance   types.InstanceConfig
	iamService generic.Service
}

// NewAWSFactory creates a new AWS factory
func NewAWSFactory(instance types.InstanceConfig) *AWSFactory {
	ctx := context.Background()

	// Create IAM service once and cache it
	iamService, err := iam.NewAWSIAMService(ctx)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to create AWS IAM service: %v\n", err)
	}

	return &AWSFactory{
		ctx:        ctx,
		instance:   instance,
		iamService: iamService,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *AWSFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	cloudParams := f.instance.CloudParams()

	switch serviceID {
	case "iam":
		if f.iamService == nil {
			return nil, fmt.Errorf("AWS IAM service not initialized")
		}
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewAWSS3Service(f.ctx, f.instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS service '%s': %w", serviceID, err)
		}
		if err := service.ElevateAccessForInspection(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to elevate access for %s: %v\n", serviceID, err)
		}
		return service, nil

	case "logging":
		service, err := logging.NewAWSLoggingService(f.ctx, &cloudParams, nil, f.instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS logging service: %w", err)
		}
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for AWS: %s", serviceID)
	}
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AWSFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity, testAccess bool) (generic.Service, error) {
	if identity.Provider != string(ProviderAWS) {
		return nil, fmt.Errorf("identity is not for AWS provider: %s", identity.Provider)
	}

	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewAWSS3ServiceWithCredentials(f.ctx, f.instance, identity)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS service '%s' with identity: %w", serviceID, err)
		}
		if err := service.ElevateAccessForInspection(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to elevate access for %s: %v\n", serviceID, err)
		}
		if testAccess {
			if err = waitForUserProvisioning(service); err != nil {
				return nil, fmt.Errorf("user provisioning validation failed: %w", err)
			}
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
