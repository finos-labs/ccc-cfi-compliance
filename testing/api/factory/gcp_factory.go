package factory

import (
	"context"
	"fmt"
	"os"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// GCPFactory implements the Factory interface for GCP
type GCPFactory struct {
	ctx         context.Context
	cloudParams environment.CloudParams
	iamService  generic.Service
}

// NewGCPFactory creates a new GCP factory
func NewGCPFactory() *GCPFactory {
	ctx := context.Background()

	// Get project ID from environment
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		projectID = os.Getenv("TF_VAR_gcp_project_id")
	}

	// Get region from environment
	region := os.Getenv("GCP_REGION")
	if region == "" {
		region = os.Getenv("TF_VAR_gcp_region")
	}
	if region == "" {
		region = "us-central1"
	}

	cloudParams := environment.CloudParams{
		Provider:     "gcp",
		Region:       region,
		GCPProjectID: projectID,
	}

	// Create IAM service once and cache it
	var iamService generic.Service
	if projectID != "" {
		var err error
		iamService, err = iam.NewGCPIAMService(ctx, projectID)
		if err != nil {
			// Log error but don't fail - IAM service might not be needed
			fmt.Printf("⚠️  Warning: Failed to create GCP IAM service: %v\n", err)
		}
	}

	return &GCPFactory{
		ctx:         ctx,
		cloudParams: cloudParams,
		iamService:  iamService,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *GCPFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	switch serviceID {
	case "iam":
		return f.iamService, nil

	case "object-storage":
		return objstorage.NewGCPStorageService(f.ctx, f.cloudParams)

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
		// IAM service doesn't need per-identity clients, return the cached IAM service
		return f.iamService, nil

	case "object-storage":
		service, err := objstorage.NewGCPStorageServiceWithCredentials(f.ctx, f.cloudParams, identity)
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
