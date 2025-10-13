package factory

import (
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
)

// GCPFactory implements the Factory interface for GCP
type GCPFactory struct {
	projectID string
	config    map[string]interface{}
}

// NewGCPFactory creates a new GCP factory
func NewGCPFactory() *GCPFactory {
	return &GCPFactory{
		config: make(map[string]interface{}),
	}
}

// GetServiceAPI returns a generic service API client for the given GCP resource name
func (f *GCPFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	// TODO: Implement GCP service client creation
	return nil, fmt.Errorf("GetServiceAPI not yet implemented for GCP: %s", serviceID)
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *GCPFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	// TODO: Implement GCP service client creation with specific identity
	if identity.Provider != string(ProviderGCP) {
		return nil, fmt.Errorf("identity is not for GCP provider: %s", identity.Provider)
	}
	return nil, fmt.Errorf("GetServiceAPIWithIdentity not yet implemented for GCP: %s", serviceID)
}

// GetIAMService returns the IAM service for GCP
func (f *GCPFactory) GetIAMService() (iam.IAMService, error) {
	// TODO: Implement GCP IAM service
	return nil, fmt.Errorf("GetIAMService not yet implemented for GCP")
}

// GetObjectStorageService returns a GCS service for the given service ID
func (f *GCPFactory) GetObjectStorageService(serviceID string) (objstorage.Service, error) {
	// TODO: Implement GCS service creation
	return nil, fmt.Errorf("GetObjectStorageService not yet implemented for GCP: %s", serviceID)
}

// GetObjectStorageServiceWithIdentity returns a GCS service authenticated as the given identity
func (f *GCPFactory) GetObjectStorageServiceWithIdentity(serviceID string, identity *iam.Identity) (objstorage.Service, error) {
	// TODO: Implement GCS service creation with specific identity
	if identity.Provider != string(ProviderGCP) {
		return nil, fmt.Errorf("identity is not for GCP provider: %s", identity.Provider)
	}
	return nil, fmt.Errorf("GetObjectStorageServiceWithIdentity not yet implemented for GCP: %s", serviceID)
}

// GetProvider returns the cloud provider
func (f *GCPFactory) GetProvider() CloudProvider {
	return ProviderGCP
}

// SetProjectID sets the GCP project ID for this factory
func (f *GCPFactory) SetProjectID(projectID string) {
	f.projectID = projectID
}
