package factory

import (
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

// CloudProvider represents the supported cloud providers
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderAzure CloudProvider = "azure"
	ProviderGCP   CloudProvider = "gcp"
)

// Factory creates cloud service API clients for different providers
type Factory interface {
	// GetServiceAPI returns a generic service API client for the given service ID
	GetServiceAPI(serviceID string) (generic.Service, error)

	// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
	GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error)

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
