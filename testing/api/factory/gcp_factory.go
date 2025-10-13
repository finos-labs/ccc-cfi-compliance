package factory

import (
	"context"
	"fmt"
	"os"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

// GCPFactory implements the Factory interface for GCP
type GCPFactory struct {
	ctx context.Context
}

// NewGCPFactory creates a new GCP factory
func NewGCPFactory() *GCPFactory {
	return &GCPFactory{
		ctx: context.Background(),
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *GCPFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	var service generic.Service
	var err error

	// Get project ID from environment
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	switch serviceID {
	case "iam":
		service, err = iam.NewGCPIAMService(f.ctx, projectID)
	case "object-storage":
		// TODO: Implement GCS service creation
		return nil, fmt.Errorf("object-storage not yet implemented for GCP")
	default:
		return nil, fmt.Errorf("unsupported service type for GCP: %s", serviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCP service '%s': %w", serviceID, err)
	}

	return service, nil
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *GCPFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	if identity.Provider != string(ProviderGCP) {
		return nil, fmt.Errorf("identity is not for GCP provider: %s", identity.Provider)
	}

	var service generic.Service
	var err error

	// Get project ID from identity or environment
	projectID := identity.Credentials["project_id"]
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT_ID")
		if projectID == "" {
			projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
	}

	switch serviceID {
	case "iam":
		// IAM service doesn't typically use per-identity clients, return the standard IAM service
		service, err = iam.NewGCPIAMService(f.ctx, projectID)

	case "object-storage":
		// TODO: Implement GCS service with credentials
		// credentialsJSON := identity.Credentials["service_account_key"]
		// service, err = objstorage.NewGCSServiceWithCredentials(f.ctx, projectID, []byte(credentialsJSON))
		return nil, fmt.Errorf("object-storage with identity not yet implemented for GCP")

	default:
		return nil, fmt.Errorf("unsupported service type for GCP: %s", serviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCP service '%s' with identity: %w", serviceID, err)
	}

	return service, nil
}

// GetProvider returns the cloud provider
func (f *GCPFactory) GetProvider() CloudProvider {
	return ProviderGCP
}

// SetContext sets the context for this factory
func (f *GCPFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}
