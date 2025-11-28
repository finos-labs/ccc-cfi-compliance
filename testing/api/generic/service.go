package generic

import (
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// Service is the generic interface for cloud services
// This interface can be extended in the future with common methods
// that all cloud services should implement
type Service interface {

	// For a given service type, return all the resources that can be tested within it,
	// as a set of TestParams.
	GetTestableResources() ([]environment.TestParams, error)
}
