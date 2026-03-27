package vpc

import "github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"

// DefaultVPC is a minimal representation of a default VPC.
// It is used for CCC.VPC controls which can be verified from control-plane metadata.
type DefaultVPC struct {
	VpcID  string
	Region string
}

// Service provides operations for VPC/networking compliance testing.
// CN01–CN04 interfaces are composed in as each control PR lands.
type Service interface {
	generic.Service
	CN01Service
	CN02Service
	TestResourceService
}
