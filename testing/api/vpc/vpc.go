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

	// EvaluateDefaultVpcControl evaluates CN01 for a VPC and returns a compact
	// structured outcome (verdict, result class, and reason).
	EvaluateDefaultVpcControl(vpcID string) (map[string]interface{}, error)

	// ListDefaultVpcs returns basic metadata for default VPCs in the configured region.
	ListDefaultVpcs() ([]DefaultVPC, error)

	// ListPublicSubnets returns a slice of objects describing subnets which are
	// considered public (have a default route to an Internet Gateway) for the
	// given VPC. Each object should include at least SubnetId and MapPublicIpOnLaunch.
	ListPublicSubnets(vpcID string) ([]interface{}, error)

	// SummarizePublicSubnets returns a human-readable summary of what will be checked
	// for CN02, including an explicit N/A marker when no public subnets are found.
	SummarizePublicSubnets(vpcID string) (string, error)

	// EvaluatePublicSubnetDefaultIPControl evaluates CN02 for a VPC and returns
	// a compact structured outcome (verdict, result class, reason, and counts).
	EvaluatePublicSubnetDefaultIPControl(vpcID string) (map[string]interface{}, error)

	// SelectPublicSubnetForTest selects one public subnet in the given VPC for
	// active/behavioral CN02 and CN04 checks.
	SelectPublicSubnetForTest(vpcID string) (map[string]interface{}, error)

	// CreateTestResourceInSubnet creates a short-lived test resource in the
	// specified subnet and returns a resource identifier.
	CreateTestResourceInSubnet(subnetID string) (map[string]interface{}, error)

	// GetResourceExternalIpAssignment reports whether the given test resource has
	// an external/public IP assigned.
	GetResourceExternalIpAssignment(resourceID string) (map[string]interface{}, error)

	// DeleteTestResource deletes a previously created test resource.
	DeleteTestResource(resourceID string) (map[string]interface{}, error)

	// EvaluatePeerAgainstAllowList classifies whether a candidate peer/requester VPC
	// appears in the configured CN03 allow-list inputs.
	EvaluatePeerAgainstAllowList(peerVpcID string) (map[string]interface{}, error)

	// AttemptVpcPeeringDryRun attempts a dry-run CreateVpcPeeringConnection from
	// requesterVpcID to peerVpcID and returns normalized evidence.
	AttemptVpcPeeringDryRun(requesterVpcID, peerVpcID string) (map[string]interface{}, error)

	// LoadVpcPeeringTrialMatrix loads a CN03 trial matrix JSON file with receiver,
	// allowed requester list, and disallowed requester list.
	LoadVpcPeeringTrialMatrix(filePath string) (map[string]interface{}, error)

	// RunVpcPeeringDryRunTrialsFromFile executes CN03 dry-run trials for all
	// requesters listed in a trial matrix JSON file.
	RunVpcPeeringDryRunTrialsFromFile(filePath string) (map[string]interface{}, error)

	// ListVpcFlowLogs returns flow log records configured for the given VPC.
	// Each returned object includes core fields used by CN04 checks.
	ListVpcFlowLogs(vpcID string) ([]interface{}, error)

	// HasActiveAllTrafficFlowLogs returns true when the VPC has at least one flow log
	// and all discovered flow logs are ACTIVE with TrafficType=ALL.
	HasActiveAllTrafficFlowLogs(vpcID string) (bool, error)

	// SummarizeVpcFlowLogs returns a human-readable CN04 summary for test evidence.
	SummarizeVpcFlowLogs(vpcID string) (string, error)

	// EvaluateVpcFlowLogsControl evaluates CN04 for a VPC and returns a compact
	// structured outcome (verdict, result class, reason, and counts).
	EvaluateVpcFlowLogsControl(vpcID string) (map[string]interface{}, error)

	// PrepareFlowLogDeliveryObservation validates CN04 observation preconditions
	// and returns setup details before active traffic generation.
	PrepareFlowLogDeliveryObservation(vpcID string) (map[string]interface{}, error)

	// GenerateTestTraffic produces best-effort short-lived traffic evidence for
	// CN04 behavioral checks.
	GenerateTestTraffic(vpcID string) (map[string]interface{}, error)

	// ObserveRecentFlowLogDelivery returns compact evidence indicating whether
	// flow-log delivery appears healthy after active checks.
	ObserveRecentFlowLogDelivery(vpcID string) (map[string]interface{}, error)
}
