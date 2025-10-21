package generic

import "github.com/finos-labs/ccc-cfi-compliance/testing/inspection"

// Service is the interface for cloud services that can enumerate their instances as TestParams
type Service interface {
	// GetAllInstances returns all instances of this service type as TestParams
	// For object storage, this would return all buckets
	// For IAM, this would return all users/identities
	GetAllInstances() ([]inspection.TestParams, error)
}
