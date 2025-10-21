package factory

import (
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

// CloudProvider represents the supported cloud providers
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderAzure CloudProvider = "azure"
	ProviderGCP   CloudProvider = "gcp"
)

// ServiceType represents the types of cloud services
type ServiceType string

const (
	ServiceTypeIAM           ServiceType = "iam"
	ServiceTypeObjectStorage ServiceType = "object-storage"
)

// Factory creates cloud service API clients for different providers
type Factory interface {
	// GetServiceAPI returns a service API client for the given service type (string version for Gherkin)
	// Returns any since the concrete service type depends on the serviceType requested
	// Callers should type-assert to the specific service interface (e.g., objstorage.Service)
	GetServiceAPI(serviceType string) (any, error)

	// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity (string version for Gherkin)
	// Returns any since the concrete service type depends on the serviceType requested
	// Callers should type-assert to the specific service interface (e.g., objstorage.Service)
	GetServiceAPIWithIdentity(serviceType string, identity *iam.Identity) (any, error)

	// GetServiceAPIForType returns a service API client for the given service type (typed version for Go code)
	// Returns any since the concrete service type depends on the ServiceType requested
	// Callers should type-assert to the specific service interface (e.g., objstorage.Service)
	GetServiceAPIForType(serviceType ServiceType) (any, error)

	// GetServiceAPIWithIdentityForType returns a service API client authenticated as the given identity (typed version for Go code)
	// Returns any since the concrete service type depends on the ServiceType requested
	// Callers should type-assert to the specific service interface (e.g., objstorage.Service)
	GetServiceAPIWithIdentityForType(serviceType ServiceType, identity *iam.Identity) (any, error)

	// GetProvider returns the cloud provider this factory is configured for
	GetProvider() CloudProvider
}

// NewFactory creates a new factory for the specified cloud provider
func NewFactory(provider CloudProvider) (Factory, error) {
	switch provider {
	case ProviderAWS:
		return NewAWSFactory(), nil
	case ProviderAzure:
		return NewAzureFactory(), nil
	case ProviderGCP:
		return NewGCPFactory(), nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}
