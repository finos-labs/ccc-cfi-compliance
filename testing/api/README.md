# Cloud Service API

This package provides a unified interface for interacting with cloud service APIs across AWS, Azure, and GCP.

## Architecture

### Factory Pattern (`factory/`)

The factory pattern provides a consistent way to create cloud service clients:

```go
// Create a factory for a specific cloud provider
factory, err := factory.NewFactory(factory.ProviderAWS)

// Get a service API client
service, err := factory.GetServiceAPI("object-storage")
service, err := factory.GetServiceAPI("iam")

// Get a service API with a specific identity
identity, err := iamService.ProvisionUser("test-user")
service, err := factory.GetServiceAPIWithIdentity("object-storage", identity)
```

### Generic Service Interface (`generic/`)

The `Service` interface provides a common abstraction for all cloud services. Currently empty but will be extended with common operations.

### IAM Service (`iam/`)

The `IAMService` interface provides identity and access management operations:

- **ProvisionUser**: Create a new user/identity
- **SetAccess**: Grant access to a service at a specific level (read/write/admin)
- **DestroyUser**: Remove an identity and all associated access

```go
// Get IAM service from factory
iamService, err := factory.GetIAMService()

// Provision a new user
identity, err := iamService.ProvisionUser("test-user")

// Grant access to a service
err = iamService.SetAccess("test-user", "service-id", iam.AccessLevelRead)

// Remove the user
err = iamService.DestroyUser("test-user")
```

## Usage in Tests

These APIs will be used by the compliance test framework to:

1. Provision test users/identities
2. Grant specific access levels to test privilege escalation
3. Interact with services using different identities
4. Clean up test resources after testing
