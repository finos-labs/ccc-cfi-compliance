package vpc

import "github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"

// DefaultVPC is a minimal representation of a default VPC.
// It is used for CCC.VPC controls which can be verified from control-plane metadata.
type DefaultVPC struct {
	VpcID  string
	Region string
}

// Service provides operations for VPC/networking compliance testing.
type Service interface {
	generic.Service

	// CountDefaultVpcs returns the number of default VPCs in the configured region.
	CountDefaultVpcs() (int, error)

	// ListDefaultVpcs returns basic metadata for default VPCs in the configured region.
	ListDefaultVpcs() ([]DefaultVPC, error)
}

