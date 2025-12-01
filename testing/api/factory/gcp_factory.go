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
	ctx        context.Context
	iamService generic.Service
}

// NewGCPFactory creates a new GCP factory
func NewGCPFactory() *GCPFactory {
	ctx := context.Background()
	
	// Get project ID from environment
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
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
		ctx:        ctx,
		iamService: iamService,
	}
}

// GetServiceAPI returns a generic service API client for the given service type
func (f *GCPFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	switch serviceID {
	case "iam":
		return f.iamService, nil
		
	case "object-storage":
		// TODO: Implement GCS service creation
		return nil, fmt.Errorf("object-storage not yet implemented for GCP")
		
	default:
		return nil, fmt.Errorf("unsupported service type for GCP: %s", serviceID)
	}
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *GCPFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	if identity.Provider != string(ProviderGCP) {
		return nil, fmt.Errorf("identity is not for GCP provider: %s", identity.Provider)
	}

	switch serviceID {
	case "iam":
		// IAM service doesn't need per-identity clients, return the cached IAM service
		return f.iamService, nil

	case "object-storage":
		// TODO: Implement GCS service with credentials
		// credentialsJSON := identity.Credentials["service_account_key"]
		// service, err = objstorage.NewGCSServiceWithCredentials(f.ctx, projectID, []byte(credentialsJSON))
		return nil, fmt.Errorf("object-storage with identity not yet implemented for GCP")

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
