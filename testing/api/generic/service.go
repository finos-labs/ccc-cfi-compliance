package generic

import (
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// Service is the generic interface for cloud services
// This interface can be extended in the future with common methods
// that all cloud services should implement
type Service interface {

	// For a given service type, return all the resources that can be tested within it,
	// as a set of TestParams. If no resources exist, create default ones.
	GetOrProvisionTestableResources() ([]environment.TestParams, error)

	// CheckUserProvisioned validates that the service's identity is properly provisioned
	// and usable. Returns nil if the user is ready, error otherwise.
	// This is used in a retry loop to ensure credentials have propagated before use.
	CheckUserProvisioned() error

	// ElevateAccessForInspection temporarily elevates access permissions to allow testing
	// For example, Azure storage might enable public network access
	// The original access level is stored internally for later reset
	ElevateAccessForInspection() error

	// ResetAccess restores the original access permissions that were in place
	// before ElevateAccessForInspection was called
	ResetAccess() error
}
