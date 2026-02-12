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

	// IsDefaultVpc reports whether the specified VPC is a "default VPC".
	IsDefaultVpc(vpcID string) (bool, error)

	// ListDefaultVpcs returns basic metadata for default VPCs in the configured region.
	ListDefaultVpcs() ([]DefaultVPC, error)

	// ListPublicSubnets returns a slice of objects describing subnets which are
	// considered public (have a default route to an Internet Gateway) for the
	// given VPC. Each object should include at least SubnetId and MapPublicIpOnLaunch.
	ListPublicSubnets(vpcID string) ([]interface{}, error)

	// SummarizePublicSubnets returns a human-readable summary of what will be checked
	// for CN02, including an explicit N/A marker when no public subnets are found.
	SummarizePublicSubnets(vpcID string) (string, error)

	// ListVpcFlowLogs returns flow log records configured for the given VPC.
	// Each returned object includes core fields used by CN04 checks.
	ListVpcFlowLogs(vpcID string) ([]interface{}, error)

	// HasActiveAllTrafficFlowLogs returns true when the VPC has at least one flow log
	// and all discovered flow logs are ACTIVE with TrafficType=ALL.
	HasActiveAllTrafficFlowLogs(vpcID string) (bool, error)

	// SummarizeVpcFlowLogs returns a human-readable CN04 summary for test evidence.
	SummarizeVpcFlowLogs(vpcID string) (string, error)
}
