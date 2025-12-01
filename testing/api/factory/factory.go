package factory

import (
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// CloudProvider represents the supported cloud providers
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderAzure CloudProvider = "azure"
	ProviderGCP   CloudProvider = "gcp"
)

// Cache for factories (one per provider)
var factoryCache = make(map[CloudProvider]Factory)

// Factory creates cloud service API clients for different providers
type Factory interface {
	// GetServiceAPI returns a generic service API client for the given service ID
	GetServiceAPI(serviceID string) (generic.Service, error)

	// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
	GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error)

	// GetProvider returns the cloud provider this factory is configured for
	GetProvider() CloudProvider
}

// NewFactory creates a new factory for the specified cloud provider using cloud-specific configuration
// Factories are cached per provider to ensure IAM service caching works across calls
func NewFactory(provider CloudProvider, cloudParams environment.CloudParams) (Factory, error) {
	// Check cache first
	if cachedFactory, exists := factoryCache[provider]; exists {
		fmt.Printf("♻️  Using cached factory for provider: %s\n", provider)
		return cachedFactory, nil
	}

	// Create new factory
	fmt.Printf("🏭 Creating new factory for provider: %s\n", provider)
	var factory Factory
	switch provider {
	case ProviderAWS:
		factory = NewAWSFactory(cloudParams)
	case ProviderAzure:
		factory = NewAzureFactory(cloudParams)
	case ProviderGCP:
		factory = NewGCPFactory()
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}

	// Cache the factory
	factoryCache[provider] = factory
	return factory, nil
}
