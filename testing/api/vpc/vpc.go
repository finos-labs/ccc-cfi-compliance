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

	// AttemptDisallowedPeeringDryRun attempts a dry-run VPC peering request to a
	// configured disallowed peer and returns normalized evidence.
	AttemptDisallowedPeeringDryRun(requesterVpcID string) (map[string]interface{}, error)

	// IsDisallowedPeeringPrevented returns true when dry-run indicates the
	// disallowed peering request was denied by policy/guardrails.
	IsDisallowedPeeringPrevented(requesterVpcID string) (bool, error)

	// EvaluateDisallowedPeeringDryRun evaluates normalized dry-run evidence and
	// returns true when the request was prevented.
	EvaluateDisallowedPeeringDryRun(evidence map[string]interface{}) (bool, error)

	// SummarizePeeringOutcomeCompact returns compact structured CN03 evidence for
	// clear visual reporting (mode, verdict, reason, key IDs, and dry-run result).
	SummarizePeeringOutcomeCompact(evidence map[string]interface{}, mode string) (map[string]interface{}, error)
}
