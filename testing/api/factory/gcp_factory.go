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

// GCPFactory implements the Factory interface for GCP
type GCPFactory struct {
	ctx        context.Context
	instance   types.InstanceConfig
	iamService generic.Service
}

// NewGCPFactory creates a new GCP factory
func NewGCPFactory(instance types.InstanceConfig) *GCPFactory {
	ctx := context.Background()
	cloudParams := instance.CloudParams()

	// Create IAM service once and cache it
	var iamService generic.Service
	if cloudParams.GcpProjectId != "" {
		var err error
		iamService, err = iam.NewGCPIAMService(ctx, instance)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to create GCP IAM service: %v\n", err)
		}
	}

	return &GCPFactory{
		ctx:        ctx,
		instance:   instance,
		iamService: iamService,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *GCPFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	cloudParams := f.instance.CloudParams()

	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		return objstorage.NewGCPStorageService(f.ctx, f.instance)

	case "logging":
		service, err := logging.NewGCPLoggingService(f.ctx, &cloudParams, nil, f.instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP logging service: %w", err)
		}
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for GCP: %s", serviceID)
	}
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *GCPFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity, testAccess bool) (generic.Service, error) {
	if identity.Provider != string(ProviderGCP) {
		return nil, fmt.Errorf("identity is not for GCP provider: %s", identity.Provider)
	}

	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewGCPStorageServiceWithCredentials(f.ctx, f.instance, identity)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCS service with credentials: %w", err)
		}
		if testAccess {
			if err := service.CheckUserProvisioned(); err != nil {
				return nil, fmt.Errorf("credentials not ready: %w", err)
			}
		}
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported service type for GCP: %s", serviceID)
	}
}

// GetProvider returns the cloud provider
func (f *GCPFactory) GetProvider() CloudProvider {
	return ProviderGCP
}

// SetContext sets the context for this factory
func (f *GCPFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}
