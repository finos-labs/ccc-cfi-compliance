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

// AzureFactory implements the Factory interface for Azure
type AzureFactory struct {
	ctx        context.Context
	instance   types.InstanceConfig
	iamService generic.Service
}

// NewAzureFactory creates a new Azure factory
func NewAzureFactory(instance types.InstanceConfig) *AzureFactory {
	ctx := context.Background()

	// Create IAM service once and cache it
	iamService, err := iam.NewAzureIAMService(ctx, instance)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to create Azure IAM service: %v\n", err)
	}

	return &AzureFactory{
		ctx:        ctx,
		instance:   instance,
		iamService: iamService,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *AzureFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	cloudParams := f.instance.CloudParams()

	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewAzureBlobService(f.ctx, cloudParams, f.instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure service '%s': %w", serviceID, err)
		}
		if err := service.ElevateAccessForInspection(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to elevate access for %s: %v\n", serviceID, err)
		}
		return service, nil

	case "logging":
		service, err := logging.NewAzureLoggingService(f.ctx, &cloudParams, nil, f.instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure logging service: %w", err)
		}
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for Azure: %s", serviceID)
	}
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AzureFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity, testAccess bool) (generic.Service, error) {
	if identity.Provider != string(ProviderAzure) {
		return nil, fmt.Errorf("identity is not for Azure provider: %s", identity.Provider)
	}

	cloudParams := f.instance.CloudParams()

	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewAzureBlobServiceWithCredentials(f.ctx, cloudParams, f.instance, identity)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure service '%s' with identity: %w", serviceID, err)
		}
		if testAccess {
			if err = waitForUserProvisioning(service); err != nil {
				return nil, fmt.Errorf("user provisioning validation failed: %w", err)
			}
		}
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for Azure: %s", serviceID)
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
