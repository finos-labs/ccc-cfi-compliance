package vpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AWSVPCService implements VPC Service for AWS EC2/VPC.
type AWSVPCService struct {
	client      *ec2.Client
	ctx         context.Context
	cloudParams environment.CloudParams
}

// NewAWSVPCService creates a new AWS VPC service using default credentials.
func NewAWSVPCService(ctx context.Context, cloudParams environment.CloudParams) (*AWSVPCService, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cloudParams.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSVPCService{
		client:      ec2.NewFromConfig(cfg),
		ctx:         ctx,
		cloudParams: cloudParams,
	}, nil
}

func (s *AWSVPCService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	// Return all VPCs in the configured region as testable resources.
	// Some controls are region-scoped, but returning per-VPC resources allows
	// controls that require a VPC ID (e.g., subnet-level checks) to execute.
	output, err := s.client.DescribeVpcs(s.ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	resources := make([]environment.TestParams, 0, len(output.Vpcs))
	for _, vpc := range output.Vpcs {
		vpcID := aws.ToString(vpc.VpcId)
		resourceName := vpcID
		if nameTag := tagValue(vpc.Tags, "Name"); nameTag != "" {
			resourceName = nameTag
		}

		resources = append(resources, environment.TestParams{
			ResourceName:        resourceName,
			UID:                 vpcID,
			ProviderServiceType: "ec2:vpc",
			ServiceType:         "vpc",
			CatalogTypes:        []string{"CCC.VPC"},
			CloudParams:         s.cloudParams,
		})
	}

	return resources, nil
}

func (s *AWSVPCService) CheckUserProvisioned() error {
	_, err := s.client.DescribeVpcs(s.ctx, &ec2.DescribeVpcsInput{MaxResults: aws.Int32(5)})
	if err != nil {
		return fmt.Errorf("credentials not ready for EC2/VPC access: %w", err)
	}
	return nil
}

func (s *AWSVPCService) ElevateAccessForInspection() error { return nil }
func (s *AWSVPCService) ResetAccess() error                { return nil }

func (s *AWSVPCService) CountDefaultVpcs() (int, error) {
	vpcs, err := s.describeDefaultVpcs()
	if err != nil {
		return 0, err
	}
	return len(vpcs), nil
}

func (s *AWSVPCService) IsDefaultVpc(vpcID string) (bool, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return false, fmt.Errorf("vpcID is required")
	}

	out, err := s.client.DescribeVpcs(s.ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcIDStr},
	})
	if err != nil {
		return false, fmt.Errorf("failed to describe vpc %s: %w", vpcIDStr, err)
	}
	if len(out.Vpcs) == 0 {
		return false, fmt.Errorf("vpc %s not found", vpcIDStr)
	}

	return aws.ToBool(out.Vpcs[0].IsDefault), nil
}

func (s *AWSVPCService) EvaluateDefaultVpcControl(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	isDefault, err := s.IsDefaultVpc(vpcIDStr)
	if err != nil {
		return nil, err
	}

	verdict := "PASS"
	resultClass := "PASS"
	compliant := true
	reason := "in-scope VPC is not default"
	if isDefault {
		verdict = "FAIL"
		resultClass = "FAIL"
		compliant = false
		reason = "in-scope VPC is default"
	}

	return map[string]interface{}{
		"Verdict":      verdict,
		"ResultClass":  resultClass,
		"Compliant":    compliant,
		"Reason":       reason,
		"VpcId":        vpcIDStr,
		"IsDefaultVpc": isDefault,
	}, nil
}

func (s *AWSVPCService) ListDefaultVpcs() ([]DefaultVPC, error) {
	vpcs, err := s.describeDefaultVpcs()
	if err != nil {
		return nil, err
	}

	out := make([]DefaultVPC, 0, len(vpcs))
	for _, vpc := range vpcs {
		out = append(out, DefaultVPC{
			VpcID:  aws.ToString(vpc.VpcId),
			Region: s.cloudParams.Region,
		})
	}
	return out, nil
}

func (s *AWSVPCService) describeDefaultVpcs() ([]types.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("is-default"),
				Values: []string{"true"},
			},
		},
	}

	resp, err := s.client.DescribeVpcs(s.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe default VPCs: %w", err)
	}
	return resp.Vpcs, nil
}

func (s *AWSVPCService) ListPublicSubnets(vpcID string) ([]interface{}, error) {
	return s.listPublicSubnets(vpcID)
}

func (s *AWSVPCService) SummarizePublicSubnets(vpcID string) (string, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return "", fmt.Errorf("vpcID is required")
	}

	outcome, err := s.EvaluatePublicSubnetDefaultIPControl(vpcIDStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"CCC.VPC.CN02.AR01: %v (%v) for VPC %s - %v",
		outcome["Verdict"],
		outcome["ResultClass"],
		vpcIDStr,
		outcome["Reason"],
	), nil
}

func (s *AWSVPCService) EvaluatePublicSubnetDefaultIPControl(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	publicSubnets, err := s.listPublicSubnets(vpcIDStr)
	if err != nil {
		return nil, err
	}

	violatingSubnetIDs := make([]string, 0)
	for _, item := range publicSubnets {
		row, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected public subnet record type")
		}
		if boolFromEvidence(row["MapPublicIpOnLaunch"]) {
			violatingSubnetIDs = append(violatingSubnetIDs, strings.TrimSpace(fmt.Sprintf("%v", row["SubnetId"])))
		}
	}

	verdict := "PASS"
	resultClass := "PASS"
	compliant := true
	reason := fmt.Sprintf("all %d public subnet(s) disable default public IP assignment", len(publicSubnets))

	if len(publicSubnets) == 0 {
		verdict = "NA"
		resultClass = "NA"
		reason = "no public subnets found for in-scope VPC"
	} else if len(violatingSubnetIDs) > 0 {
		verdict = "FAIL"
		resultClass = "FAIL"
		compliant = false
		reason = fmt.Sprintf("%d public subnet(s) have MapPublicIpOnLaunch=true", len(violatingSubnetIDs))
	}

	return map[string]interface{}{
		"Verdict":              verdict,
		"ResultClass":          resultClass,
		"Compliant":            compliant,
		"Reason":               reason,
		"VpcId":                vpcIDStr,
		"PublicSubnetCount":    len(publicSubnets),
		"ViolatingSubnetCount": len(violatingSubnetIDs),
		"ViolatingSubnetIds":   violatingSubnetIDs,
	}, nil
}

func (s *AWSVPCService) SelectPublicSubnetForTest(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", vpcID))
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	publicSubnets, err := s.listPublicSubnets(vpcIDStr)
	if err != nil {
		return nil, err
	}
	if len(publicSubnets) == 0 {
		return nil, fmt.Errorf("no public subnets found for VPC %s", vpcIDStr)
	}

	rows := make([]map[string]interface{}, 0, len(publicSubnets))
	for _, item := range publicSubnets {
		row, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected public subnet record type")
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.TrimSpace(fmt.Sprintf("%v", rows[i]["SubnetId"])) <
			strings.TrimSpace(fmt.Sprintf("%v", rows[j]["SubnetId"]))
	})

	selected := rows[0]
	return map[string]interface{}{
		"VpcId":                vpcIDStr,
		"SubnetId":             strings.TrimSpace(fmt.Sprintf("%v", selected["SubnetId"])),
		"RouteTableId":         strings.TrimSpace(fmt.Sprintf("%v", selected["RouteTableId"])),
		"MapPublicIpOnLaunch":  boolFromEvidence(selected["MapPublicIpOnLaunch"]),
		"PublicSubnetCount":    len(rows),
		"SelectionDescription": "first public subnet by SubnetId",
	}, nil
}

func (s *AWSVPCService) CreateTestResourceInSubnet(subnetID string) (map[string]interface{}, error) {
	subnetIDStr := strings.TrimSpace(fmt.Sprintf("%v", subnetID))
	if subnetIDStr == "" {
		return nil, fmt.Errorf("subnetID is required")
	}

	amiID := cnTestAmiID()
	if amiID == "" {
		return nil, fmt.Errorf("missing test AMI input: set CN_TEST_AMI_ID or CN02_TEST_AMI_ID")
	}

	instanceType := cnTestInstanceType()
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: types.InstanceType(instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		SubnetId:     aws.String(subnetIDStr),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String("cfi-vpc-test-resource")},
					{Key: aws.String("ManagedBy"), Value: aws.String("CCC-CFI-Compliance")},
					{Key: aws.String("CFIControlSet"), Value: aws.String("CCC.VPC")},
					{Key: aws.String("CFITest"), Value: aws.String("true")},
				},
			},
		},
	}

	out, err := s.client.RunInstances(s.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create test resource in subnet %s: %w", subnetIDStr, err)
	}
	if len(out.Instances) == 0 {
		return nil, fmt.Errorf("failed to create test resource in subnet %s: empty RunInstances response", subnetIDStr)
	}

	resourceID := strings.TrimSpace(aws.ToString(out.Instances[0].InstanceId))
	if resourceID == "" {
		return nil, fmt.Errorf("failed to create test resource in subnet %s: missing instance id", subnetIDStr)
	}

	// Best-effort wait so subsequent describe calls have stable state.
	_ = s.waitForInstanceTerminalOrRunning(resourceID, 2*time.Minute)

	return map[string]interface{}{
		"ResourceId":   resourceID,
		"ResourceType": "ec2:instance",
		"SubnetId":     subnetIDStr,
		"AmiId":        amiID,
		"InstanceType": instanceType,
	}, nil
}

func (s *AWSVPCService) GetResourceExternalIpAssignment(resourceID string) (map[string]interface{}, error) {
	resourceIDStr := strings.TrimSpace(fmt.Sprintf("%v", resourceID))
	if resourceIDStr == "" {
		return nil, fmt.Errorf("resourceID is required")
	}

	instance, err := s.describeInstance(resourceIDStr)
	if err != nil {
		return nil, err
	}

	publicIP := strings.TrimSpace(aws.ToString(instance.PublicIpAddress))
	return map[string]interface{}{
		"ResourceId":    resourceIDStr,
		"ResourceType":  "ec2:instance",
		"HasExternalIp": publicIP != "",
		"ExternalIp":    publicIP,
		"State":         string(instance.State.Name),
		"VpcId":         strings.TrimSpace(aws.ToString(instance.VpcId)),
		"SubnetId":      strings.TrimSpace(aws.ToString(instance.SubnetId)),
	}, nil
}

func (s *AWSVPCService) DeleteTestResource(resourceID string) (map[string]interface{}, error) {
	resourceIDStr := strings.TrimSpace(fmt.Sprintf("%v", resourceID))
	if resourceIDStr == "" {
		return nil, fmt.Errorf("resourceID is required")
	}

	_, err := s.client.TerminateInstances(s.ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{resourceIDStr},
	})
	if err != nil {
		if isEC2NotFoundError(err) {
			return map[string]interface{}{
				"ResourceId": resourceIDStr,
				"Deleted":    true,
				"Reason":     "resource already absent",
			}, nil
		}
		return nil, fmt.Errorf("failed to delete test resource %s: %w", resourceIDStr, err)
	}

	waitTimeout := cnTestDeleteWaitTimeout()
	if waitTimeout <= 0 {
		return map[string]interface{}{
			"ResourceId":    resourceIDStr,
			"Deleted":       true,
			"CleanupStatus": "termination-requested",
			"Reason":        "async cleanup requested; termination continues in AWS control plane",
		}, nil
	}

	waitErr := s.waitForInstanceTermination(resourceIDStr, waitTimeout)
	if waitErr != nil {
		return map[string]interface{}{
			"ResourceId":    resourceIDStr,
			"Deleted":       false,
			"CleanupStatus": "termination-requested",
			"Reason":        waitErr.Error(),
		}, nil
	}

	return map[string]interface{}{
		"ResourceId": resourceIDStr,
		"Deleted":    true,
		"Reason":     "terminated",
	}, nil
}

func (s *AWSVPCService) EvaluatePeerAgainstAllowList(peerVpcID string) (map[string]interface{}, error) {
	peerVpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", peerVpcID))
	if peerVpcIDStr == "" {
		return nil, fmt.Errorf("peerVpcID is required")
	}

	allowedIDs, source, err := s.resolveCN03AllowedRequesterVpcIDs()
	if err != nil {
		return nil, err
	}

	allowed := false
	for _, allowedID := range allowedIDs {
		if allowedID == peerVpcIDStr {
			allowed = true
			break
		}
	}

	reason := "CN03 allow-list is not defined via environment variables or trial matrix file; classification is non-enforcing until IAM/SCP guardrail is configured"
	if len(allowedIDs) > 0 {
		if allowed {
			reason = "requester VPC exists in CN03 allow-list; expected enforcement outcome is allow"
		} else {
			reason = "requester VPC does not exist in CN03 allow-list; expected enforcement outcome is deny"
		}
	}

	return map[string]interface{}{
		"PeerVpcId":              peerVpcIDStr,
		"Allowed":                allowed,
		"AllowedListDefined":     len(allowedIDs) > 0,
		"AllowedListCount":       len(allowedIDs),
		"AllowedRequesterVpcIds": allowedIDs,
		"AllowedPeerVpcIds":      allowedIDs,
		"AllowListSource":        source,
		"Reason":                 reason,
	}, nil
}

func (s *AWSVPCService) AttemptVpcPeeringDryRun(requesterVpcID, peerVpcID string) (map[string]interface{}, error) {
	return s.attemptVpcPeeringDryRunWithOwner(requesterVpcID, peerVpcID, cn03PeerOwnerID())
}

func (s *AWSVPCService) LoadVpcPeeringTrialMatrix(filePath string) (map[string]interface{}, error) {
	matrix, resolvedPath, err := s.loadCN03TrialMatrix(filePath)
	if err != nil {
		return nil, err
	}

	allRequesterIDs := make([]string, 0, len(matrix.AllowedRequesterVpcIDs)+len(matrix.DisallowedRequesterVpcIDs))
	allRequesterIDs = append(allRequesterIDs, matrix.AllowedRequesterVpcIDs...)
	allRequesterIDs = append(allRequesterIDs, matrix.DisallowedRequesterVpcIDs...)

	return map[string]interface{}{
		"FilePath":                   resolvedPath,
		"ReceiverVpcId":              matrix.ReceiverVpcID,
		"PeerOwnerId":                matrix.PeerOwnerID,
		"AllowedRequesterVpcIds":     matrix.AllowedRequesterVpcIDs,
		"DisallowedRequesterVpcIds":  matrix.DisallowedRequesterVpcIDs,
		"AllowedPeerVpcIds":          matrix.AllowedRequesterVpcIDs,
		"DisallowedPeerVpcIds":       matrix.DisallowedRequesterVpcIDs,
		"AllowedCount":               len(matrix.AllowedRequesterVpcIDs),
		"DisallowedCount":            len(matrix.DisallowedRequesterVpcIDs),
		"AllRequesterVpcIds":         allRequesterIDs,
		"AllRequesterCount":          len(allRequesterIDs),
		"AllowedListDefined":         len(matrix.AllowedRequesterVpcIDs) > 0,
		"DisallowedListDefined":      len(matrix.DisallowedRequesterVpcIDs) > 0,
		"ReceiverVpcIdMatchesRegion": matrix.ReceiverVpcID != "",
	}, nil
}

func (s *AWSVPCService) RunVpcPeeringDryRunTrialsFromFile(filePath string) (map[string]interface{}, error) {
	matrix, resolvedPath, err := s.loadCN03TrialMatrix(filePath)
	if err != nil {
		return nil, err
	}
	if matrix.ReceiverVpcID == "" {
		return nil, fmt.Errorf("CN03 trial matrix file %s is missing receiver_vpc_id", resolvedPath)
	}

	trials := make([]interface{}, 0, len(matrix.AllowedRequesterVpcIDs)+len(matrix.DisallowedRequesterVpcIDs))
	unexpectedCount := 0

	runTrials := func(requesterIDs []string, expectedAllowed bool) error {
		for _, requesterID := range requesterIDs {
			evidence, dryRunErr := s.attemptVpcPeeringDryRunWithOwner(requesterID, matrix.ReceiverVpcID, matrix.PeerOwnerID)
			if dryRunErr != nil {
				return dryRunErr
			}

			actualAllowed := boolFromEvidence(evidence["DryRunAllowed"])
			matchesExpectation := actualAllowed == expectedAllowed
			if !matchesExpectation {
				unexpectedCount++
			}

			trial := map[string]interface{}{
				"RequesterVpcId":     requesterID,
				"ReceiverVpcId":      matrix.ReceiverVpcID,
				"ExpectedAllowed":    expectedAllowed,
				"DryRunAllowed":      actualAllowed,
				"MatchesExpectation": matchesExpectation,
				"ExitCode":           evidence["ExitCode"],
				"ErrorCode":          evidence["ErrorCode"],
				"Stderr":             evidence["Stderr"],
			}

			trials = append(trials, trial)
		}
		return nil
	}

	if err := runTrials(matrix.DisallowedRequesterVpcIDs, false); err != nil {
		return nil, err
	}
	if err := runTrials(matrix.AllowedRequesterVpcIDs, true); err != nil {
		return nil, err
	}

	totalTrials := len(trials)
	return map[string]interface{}{
		"FilePath":                  resolvedPath,
		"ReceiverVpcId":             matrix.ReceiverVpcID,
		"PeerOwnerId":               matrix.PeerOwnerID,
		"AllowedRequesterVpcIds":    matrix.AllowedRequesterVpcIDs,
		"DisallowedRequesterVpcIds": matrix.DisallowedRequesterVpcIDs,
		"AllowedCount":              len(matrix.AllowedRequesterVpcIDs),
		"DisallowedCount":           len(matrix.DisallowedRequesterVpcIDs),
		"TotalTrials":               totalTrials,
		"UnexpectedCount":           unexpectedCount,
		"Compliant":                 totalTrials > 0 && unexpectedCount == 0,
		"Trials":                    trials,
	}, nil
}

func (s *AWSVPCService) ListVpcFlowLogs(vpcID string) ([]interface{}, error) {
	return s.listVpcFlowLogs(vpcID)
}

func (s *AWSVPCService) HasActiveAllTrafficFlowLogs(vpcID string) (bool, error) {
	outcome, err := s.EvaluateVpcFlowLogsControl(vpcID)
	if err != nil {
		return false, err
	}
	return boolFromEvidence(outcome["Compliant"]), nil
}

func (s *AWSVPCService) SummarizeVpcFlowLogs(vpcID string) (string, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return "", fmt.Errorf("vpcID is required")
	}

	outcome, err := s.EvaluateVpcFlowLogsControl(vpcIDStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"CCC.VPC.CN04.AR01: %v (%v) for VPC %s - %v",
		outcome["Verdict"],
		outcome["ResultClass"],
		vpcIDStr,
		outcome["Reason"],
	), nil
}

func (s *AWSVPCService) EvaluateVpcFlowLogsControl(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	flowLogs, err := s.listVpcFlowLogs(vpcIDStr)
	if err != nil {
		return nil, err
	}

	nonCompliantFlowLogIDs := make([]string, 0)
	for _, item := range flowLogs {
		row, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected flow log record type")
		}

		status := strings.TrimSpace(fmt.Sprintf("%v", row["FlowLogStatus"]))
		trafficType := strings.TrimSpace(fmt.Sprintf("%v", row["TrafficType"]))
		if status != "ACTIVE" || trafficType != "ALL" {
			nonCompliantFlowLogIDs = append(nonCompliantFlowLogIDs, strings.TrimSpace(fmt.Sprintf("%v", row["FlowLogId"])))
		}
	}

	verdict := "PASS"
	resultClass := "PASS"
	compliant := true
	reason := fmt.Sprintf("all %d VPC flow log(s) are ACTIVE with TrafficType=ALL", len(flowLogs))

	if len(flowLogs) == 0 {
		verdict = "FAIL"
		resultClass = "FAIL"
		compliant = false
		reason = "no VPC flow logs are configured"
	} else if len(nonCompliantFlowLogIDs) > 0 {
		verdict = "FAIL"
		resultClass = "FAIL"
		compliant = false
		reason = fmt.Sprintf("%d flow log(s) are not ACTIVE and TrafficType=ALL", len(nonCompliantFlowLogIDs))
	}

	return map[string]interface{}{
		"Verdict":                verdict,
		"ResultClass":            resultClass,
		"Compliant":              compliant,
		"Reason":                 reason,
		"VpcId":                  vpcIDStr,
		"FlowLogCount":           len(flowLogs),
		"NonCompliantFlowLogIds": nonCompliantFlowLogIDs,
		"NonCompliantCount":      len(nonCompliantFlowLogIDs),
	}, nil
}

func (s *AWSVPCService) PrepareFlowLogDeliveryObservation(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", vpcID))
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	flowLogs, err := s.listVpcFlowLogs(vpcIDStr)
	if err != nil {
		return nil, err
	}

	activeAllCount := 0
	deliverySuccessCount := 0
	for _, item := range flowLogs {
		row, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected flow log record type")
		}

		status := strings.TrimSpace(fmt.Sprintf("%v", row["FlowLogStatus"]))
		trafficType := strings.TrimSpace(fmt.Sprintf("%v", row["TrafficType"]))
		deliverStatus := strings.TrimSpace(fmt.Sprintf("%v", row["DeliverLogsStatus"]))
		if status == "ACTIVE" && trafficType == "ALL" {
			activeAllCount++
		}
		if strings.EqualFold(deliverStatus, "SUCCESS") {
			deliverySuccessCount++
		}
	}

	ready := len(flowLogs) > 0 && activeAllCount == len(flowLogs)
	return map[string]interface{}{
		"VpcId":                vpcIDStr,
		"FlowLogCount":         len(flowLogs),
		"ActiveAllCount":       activeAllCount,
		"DeliverySuccessCount": deliverySuccessCount,
		"Ready":                ready,
		"Reason":               "flow-log preconditions evaluated for behavioral observation",
	}, nil
}

func (s *AWSVPCService) GenerateTestTraffic(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", vpcID))
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	subnetSelection, err := s.SelectPublicSubnetForTest(vpcIDStr)
	if err != nil {
		return nil, err
	}
	subnetID := strings.TrimSpace(fmt.Sprintf("%v", subnetSelection["SubnetId"]))
	if subnetID == "" {
		return nil, fmt.Errorf("no subnet selected for VPC %s", vpcIDStr)
	}

	resource, err := s.CreateTestResourceInSubnet(subnetID)
	if err != nil {
		return nil, err
	}
	resourceID := strings.TrimSpace(fmt.Sprintf("%v", resource["ResourceId"]))
	if resourceID == "" {
		return nil, fmt.Errorf("test resource creation did not return ResourceId")
	}

	// Give the launched resource a brief window to emit baseline network events.
	time.Sleep(10 * time.Second)

	externalIPEvidence, inspectErr := s.GetResourceExternalIpAssignment(resourceID)
	cleanupResult, cleanupErr := s.DeleteTestResource(resourceID)

	out := map[string]interface{}{
		"VpcId":        vpcIDStr,
		"SubnetId":     subnetID,
		"Generated":    true,
		"ResourceId":   resourceID,
		"ResourceType": "ec2:instance",
	}

	if inspectErr != nil {
		out["InspectionError"] = inspectErr.Error()
	} else {
		out["HasExternalIp"] = boolFromEvidence(externalIPEvidence["HasExternalIp"])
		out["ExternalIp"] = strings.TrimSpace(fmt.Sprintf("%v", externalIPEvidence["ExternalIp"]))
	}

	if cleanupErr != nil {
		out["CleanupError"] = cleanupErr.Error()
		out["CleanupDeleted"] = false
	} else {
		out["CleanupDeleted"] = boolFromEvidence(cleanupResult["Deleted"])
	}

	return out, nil
}

func (s *AWSVPCService) ObserveRecentFlowLogDelivery(vpcID string) (map[string]interface{}, error) {
	vpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", vpcID))
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	flowLogs, err := s.listVpcFlowLogs(vpcIDStr)
	if err != nil {
		return nil, err
	}

	deliverySuccessCount := 0
	for _, item := range flowLogs {
		row, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected flow log record type")
		}

		status := strings.TrimSpace(fmt.Sprintf("%v", row["FlowLogStatus"]))
		deliverStatus := strings.TrimSpace(fmt.Sprintf("%v", row["DeliverLogsStatus"]))
		if status == "ACTIVE" && strings.EqualFold(deliverStatus, "SUCCESS") {
			deliverySuccessCount++
		}
	}

	recordsObserved := deliverySuccessCount > 0
	reason := "no ACTIVE flow logs with DeliverLogsStatus=SUCCESS detected"
	if recordsObserved {
		reason = "at least one ACTIVE flow log reports DeliverLogsStatus=SUCCESS"
	}

	return map[string]interface{}{
		"VpcId":                vpcIDStr,
		"FlowLogCount":         len(flowLogs),
		"DeliverySuccessCount": deliverySuccessCount,
		"RecordsObserved":      recordsObserved,
		"Reason":               reason,
	}, nil
}

func (s *AWSVPCService) listPublicSubnets(vpcID string) ([]interface{}, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	subnetsOut, err := s.client.DescribeSubnets(s.ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcIDStr},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	publicSubnets := make([]interface{}, 0)
	for _, subnet := range subnetsOut.Subnets {
		subnetID := aws.ToString(subnet.SubnetId)
		if subnetID == "" {
			continue
		}

		isPublic, routeTableID, err := s.isSubnetPublic(vpcIDStr, subnetID)
		if err != nil {
			return nil, err
		}
		if !isPublic {
			continue
		}

		publicSubnets = append(publicSubnets, map[string]interface{}{
			"VpcId":               vpcIDStr,
			"SubnetId":            subnetID,
			"RouteTableId":        routeTableID,
			"MapPublicIpOnLaunch": aws.ToBool(subnet.MapPublicIpOnLaunch),
		})
	}

	return publicSubnets, nil
}

func (s *AWSVPCService) listVpcFlowLogs(vpcID string) ([]interface{}, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return nil, fmt.Errorf("vpcID is required")
	}

	out, err := s.client.DescribeFlowLogs(s.ctx, &ec2.DescribeFlowLogsInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{vpcIDStr},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe flow logs for vpc %s: %w", vpcIDStr, err)
	}

	flowLogs := make([]interface{}, 0, len(out.FlowLogs))
	for _, fl := range out.FlowLogs {
		flowLogs = append(flowLogs, map[string]interface{}{
			"VpcId":              vpcIDStr,
			"FlowLogId":          aws.ToString(fl.FlowLogId),
			"FlowLogStatus":      aws.ToString(fl.FlowLogStatus),
			"TrafficType":        string(fl.TrafficType),
			"LogDestinationType": string(fl.LogDestinationType),
			"LogDestination":     aws.ToString(fl.LogDestination),
			"DeliverLogsStatus":  aws.ToString(fl.DeliverLogsStatus),
			"DeliverLogsError":   aws.ToString(fl.DeliverLogsErrorMessage),
		})
	}

	return flowLogs, nil
}

func (s *AWSVPCService) isSubnetPublic(vpcID, subnetID string) (bool, string, error) {
	// First, look for a route table explicitly associated to the subnet.
	rtOut, err := s.client.DescribeRouteTables(s.ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("association.subnet-id"),
				Values: []string{subnetID},
			},
		},
	})
	if err != nil {
		return false, "", fmt.Errorf("failed to describe route tables for subnet %s: %w", subnetID, err)
	}

	var routeTables []types.RouteTable
	if len(rtOut.RouteTables) > 0 {
		routeTables = rtOut.RouteTables
	} else {
		// Fall back to the main route table for the VPC.
		mainOut, err := s.client.DescribeRouteTables(s.ctx, &ec2.DescribeRouteTablesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
				{
					Name:   aws.String("association.main"),
					Values: []string{"true"},
				},
			},
		})
		if err != nil {
			return false, "", fmt.Errorf("failed to describe main route table for vpc %s: %w", vpcID, err)
		}
		routeTables = mainOut.RouteTables
	}

	for _, rt := range routeTables {
		routeTableID := aws.ToString(rt.RouteTableId)
		for _, route := range rt.Routes {
			if aws.ToString(route.DestinationCidrBlock) != "0.0.0.0/0" {
				continue
			}
			gw := aws.ToString(route.GatewayId)
			if len(gw) > 4 && gw[:4] == "igw-" {
				return true, routeTableID, nil
			}
		}
		return false, routeTableID, nil
	}

	return false, "", nil
}

type cn03TrialMatrix struct {
	ReceiverVpcID             string
	PeerOwnerID               string
	AllowedRequesterVpcIDs    []string
	DisallowedRequesterVpcIDs []string
}

func (s *AWSVPCService) attemptVpcPeeringDryRunWithOwner(requesterVpcID, peerVpcID, peerOwnerID string) (map[string]interface{}, error) {
	requesterVpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", requesterVpcID))
	peerVpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", peerVpcID))
	peerOwnerIDStr := strings.TrimSpace(fmt.Sprintf("%v", peerOwnerID))

	if requesterVpcIDStr == "" {
		return nil, fmt.Errorf("requesterVpcID is required")
	}
	if peerVpcIDStr == "" {
		return nil, fmt.Errorf("peerVpcID is required")
	}

	input := &ec2.CreateVpcPeeringConnectionInput{
		VpcId:     aws.String(requesterVpcIDStr),
		PeerVpcId: aws.String(peerVpcIDStr),
		DryRun:    aws.Bool(true),
	}
	if peerOwnerIDStr != "" {
		input.PeerOwnerId = aws.String(peerOwnerIDStr)
	}

	evidence := map[string]interface{}{
		"RequesterVpcId": requesterVpcIDStr,
		"PeerVpcId":      peerVpcIDStr,
		"ReceiverVpcId":  peerVpcIDStr,
		"PeerOwnerId":    peerOwnerIDStr,
		"DryRunAllowed":  false,
		"ExitCode":       1,
		"ErrorCode":      "",
		"Stderr":         "",
		"Reason":         "request denied",
	}

	_, err := s.client.CreateVpcPeeringConnection(s.ctx, input)
	if err == nil {
		evidence["DryRunAllowed"] = true
		evidence["ExitCode"] = 0
		evidence["Reason"] = "dry-run call returned success; request would be allowed"
		return s.enrichCN03EnforcementEvidence(requesterVpcIDStr, evidence), nil
	}

	errText := strings.TrimSpace(err.Error())
	evidence["Stderr"] = errText
	evidence["Reason"] = errText

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		errorCode := strings.TrimSpace(apiErr.ErrorCode())
		evidence["ErrorCode"] = errorCode

		if strings.EqualFold(errorCode, "DryRunOperation") {
			evidence["DryRunAllowed"] = true
			evidence["ExitCode"] = 0
			evidence["Reason"] = "DryRunOperation indicates request would be allowed"
		}
		return s.enrichCN03EnforcementEvidence(requesterVpcIDStr, evidence), nil
	}

	if strings.Contains(strings.ToLower(errText), "dryrunoperation") {
		evidence["DryRunAllowed"] = true
		evidence["ExitCode"] = 0
		evidence["ErrorCode"] = "DryRunOperation"
		evidence["Reason"] = "dry-run response indicates request would be allowed"
	}

	return s.enrichCN03EnforcementEvidence(requesterVpcIDStr, evidence), nil
}

func (s *AWSVPCService) enrichCN03EnforcementEvidence(requesterVpcID string, evidence map[string]interface{}) map[string]interface{} {
	allowedIDs, source, err := s.resolveCN03AllowedRequesterVpcIDs()
	if err != nil {
		evidence["AllowListDefined"] = false
		evidence["AllowListSource"] = ""
		evidence["RequesterInAllowList"] = false
		evidence["GuardrailExpectation"] = ""
		evidence["GuardrailMismatch"] = false
		evidence["Reason"] = fmt.Sprintf("%v; CN03 allow-list resolution failed: %v", evidence["Reason"], err)
		return evidence
	}

	allowListDefined := len(allowedIDs) > 0
	requesterInAllowList := false
	for _, allowedID := range allowedIDs {
		if allowedID == requesterVpcID {
			requesterInAllowList = true
			break
		}
	}

	evidence["AllowListDefined"] = allowListDefined
	evidence["AllowListSource"] = source
	evidence["RequesterInAllowList"] = requesterInAllowList

	if !allowListDefined {
		evidence["GuardrailExpectation"] = ""
		evidence["GuardrailMismatch"] = false
		evidence["Reason"] = fmt.Sprintf("%v; CN03 allow-list is not defined, so enforcement expectation cannot be computed", evidence["Reason"])
		return evidence
	}

	expectedAllowed := requesterInAllowList
	guardrailExpectation := "deny"
	if expectedAllowed {
		guardrailExpectation = "allow"
	}

	actualAllowed := boolFromEvidence(evidence["DryRunAllowed"])
	guardrailMismatch := actualAllowed != expectedAllowed

	evidence["GuardrailExpectation"] = guardrailExpectation
	evidence["GuardrailMismatch"] = guardrailMismatch
	if guardrailMismatch {
		evidence["Reason"] = fmt.Sprintf("%v; CN03 guardrail mismatch: allow-list expects %s for requester %s", evidence["Reason"], guardrailExpectation, requesterVpcID)
	} else {
		evidence["Reason"] = fmt.Sprintf("%v; CN03 guardrail aligned: allow-list expects %s for requester %s", evidence["Reason"], guardrailExpectation, requesterVpcID)
	}

	return evidence
}

func (s *AWSVPCService) resolveCN03AllowedRequesterVpcIDs() ([]string, string, error) {
	if ids := normalizeStringList([]string{os.Getenv("CN03_ALLOWED_REQUESTER_VPC_IDS")}); len(ids) > 0 {
		return ids, "CN03_ALLOWED_REQUESTER_VPC_IDS", nil
	}
	if ids := cn03IndexedEnvValues("CN03_ALLOWED_REQUESTER_VPC_ID_"); len(ids) > 0 {
		return ids, "CN03_ALLOWED_REQUESTER_VPC_ID_1..N", nil
	}
	if ids := normalizeStringList([]string{os.Getenv("CN03_ALLOWED_PEER_VPC_IDS")}); len(ids) > 0 {
		return ids, "CN03_ALLOWED_PEER_VPC_IDS", nil
	}
	if ids := cn03IndexedEnvValues("CN03_ALLOWED_PEER_VPC_ID_"); len(ids) > 0 {
		return ids, "CN03_ALLOWED_PEER_VPC_ID_1..N", nil
	}

	matrixPath := strings.TrimSpace(os.Getenv("CN03_PEER_TRIAL_MATRIX_FILE"))
	if matrixPath != "" {
		matrix, resolvedPath, err := s.loadCN03TrialMatrix(matrixPath)
		if err != nil {
			return nil, "", err
		}
		if len(matrix.AllowedRequesterVpcIDs) > 0 {
			return matrix.AllowedRequesterVpcIDs, fmt.Sprintf("CN03_PEER_TRIAL_MATRIX_FILE (%s)", resolvedPath), nil
		}
	}

	return []string{}, "", nil
}

func (s *AWSVPCService) loadCN03TrialMatrix(filePath string) (cn03TrialMatrix, string, error) {
	resolvedPath := strings.TrimSpace(filePath)
	if resolvedPath == "" {
		resolvedPath = strings.TrimSpace(os.Getenv("CN03_PEER_TRIAL_MATRIX_FILE"))
	}
	if resolvedPath == "" {
		return cn03TrialMatrix{}, "", fmt.Errorf("filePath is required (or set CN03_PEER_TRIAL_MATRIX_FILE)")
	}

	if !filepath.IsAbs(resolvedPath) {
		if absPath, absErr := filepath.Abs(resolvedPath); absErr == nil {
			resolvedPath = absPath
		}
	}

	fileData, err := os.ReadFile(resolvedPath)
	if err != nil {
		return cn03TrialMatrix{}, "", fmt.Errorf("failed to read CN03 trial matrix file %s: %w", resolvedPath, err)
	}

	raw := make(map[string]interface{})
	if err := json.Unmarshal(fileData, &raw); err != nil {
		return cn03TrialMatrix{}, "", fmt.Errorf("failed to parse CN03 trial matrix file %s: %w", resolvedPath, err)
	}

	receiverVpcID := firstNonEmptyString(
		cn03String(raw["receiver_vpc_id"]),
		cn03String(raw["peer_vpc_id"]),
		cn03String(raw["target_vpc_id"]),
		cn03String(raw["receiverVpcId"]),
		cn03String(raw["peerVpcId"]),
		cn03String(raw["targetVpcId"]),
	)

	allowedRequesterIDs := make([]string, 0)
	allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(raw["allowed_requester_vpc_ids"])...)
	allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(raw["allowed_peer_vpc_ids"])...)
	allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(raw["allowed_vpc_ids"])...)

	disallowedRequesterIDs := make([]string, 0)
	disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(raw["disallowed_requester_vpc_ids"])...)
	disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(raw["disallowed_peer_vpc_ids"])...)
	disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(raw["disallowed_vpc_ids"])...)

	if requesters, ok := raw["requesters"].(map[string]interface{}); ok {
		allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(requesters["allowed"])...)
		disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(requesters["disallowed"])...)
		receiverVpcID = firstNonEmptyString(
			receiverVpcID,
			cn03String(requesters["receiver_vpc_id"]),
			cn03String(requesters["receiverVpcId"]),
		)
	}

	allowedRequesterIDs = normalizeStringList(allowedRequesterIDs)
	disallowedRequesterIDs = normalizeStringList(disallowedRequesterIDs)
	if len(allowedRequesterIDs) == 0 && len(disallowedRequesterIDs) == 0 {
		return cn03TrialMatrix{}, "", fmt.Errorf("CN03 trial matrix file %s does not define any requester VPC IDs", resolvedPath)
	}

	peerOwnerID := firstNonEmptyString(
		cn03String(raw["peer_owner_id"]),
		cn03String(raw["peerOwnerId"]),
		cn03PeerOwnerID(),
	)

	return cn03TrialMatrix{
		ReceiverVpcID:             receiverVpcID,
		PeerOwnerID:               peerOwnerID,
		AllowedRequesterVpcIDs:    allowedRequesterIDs,
		DisallowedRequesterVpcIDs: disallowedRequesterIDs,
	}, resolvedPath, nil
}

func (s *AWSVPCService) describeInstance(instanceID string) (types.Instance, error) {
	out, err := s.client.DescribeInstances(s.ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		if isEC2NotFoundError(err) {
			return types.Instance{}, err
		}
		return types.Instance{}, fmt.Errorf("failed to describe resource %s: %w", instanceID, err)
	}

	for _, reservation := range out.Reservations {
		for _, instance := range reservation.Instances {
			if strings.TrimSpace(aws.ToString(instance.InstanceId)) == instanceID {
				return instance, nil
			}
		}
	}

	return types.Instance{}, fmt.Errorf("resource %s not found", instanceID)
}

func (s *AWSVPCService) waitForInstanceTerminalOrRunning(instanceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		instance, err := s.describeInstance(instanceID)
		if err != nil {
			return err
		}

		switch instance.State.Name {
		case types.InstanceStateNameRunning, types.InstanceStateNameStopped, types.InstanceStateNameTerminated, types.InstanceStateNameShuttingDown:
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for instance %s to stabilize; last state=%s", instanceID, instance.State.Name)
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *AWSVPCService) waitForInstanceTermination(instanceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		instance, err := s.describeInstance(instanceID)
		if err != nil {
			if isEC2NotFoundError(err) || strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil
			}
			return err
		}

		if instance.State.Name == types.InstanceStateNameTerminated {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for instance %s termination; last state=%s", instanceID, instance.State.Name)
		}
		time.Sleep(5 * time.Second)
	}
}

func isEC2NotFoundError(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	code := strings.ToLower(strings.TrimSpace(apiErr.ErrorCode()))
	return strings.Contains(code, "notfound")
}

func tagValue(tags []types.Tag, key string) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == key {
			return aws.ToString(t.Value)
		}
	}
	return ""
}

func cnTestAmiID() string {
	for _, key := range []string{"CN_TEST_AMI_ID", "CN02_TEST_AMI_ID", "TEST_AMI_ID"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}

func cnTestInstanceType() string {
	for _, key := range []string{"CN_TEST_INSTANCE_TYPE", "CN02_TEST_INSTANCE_TYPE", "TEST_INSTANCE_TYPE"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return "t3.micro"
}

func cnTestDeleteWaitTimeout() time.Duration {
	// Default is async cleanup for faster behavioral tests.
	for _, key := range []string{"CN_TEST_DELETE_WAIT_SECONDS", "CN02_TEST_DELETE_WAIT_SECONDS"} {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}

		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}

	return 0
}

func cn03IndexedEnvValues(prefix string) []string {
	values := make([]string, 0)
	for i := 1; i <= 99; i++ {
		envKey := fmt.Sprintf("%s%d", prefix, i)
		rawValue := strings.TrimSpace(os.Getenv(envKey))
		if rawValue == "" {
			continue
		}
		values = append(values, rawValue)
	}
	return normalizeStringList(values)
}

func cn03String(value interface{}) string {
	if value == nil {
		return ""
	}
	out := strings.TrimSpace(fmt.Sprintf("%v", value))
	if out == "<nil>" {
		return ""
	}
	return out
}

func cn03StringSlice(value interface{}) []string {
	switch typedValue := value.(type) {
	case nil:
		return []string{}
	case string:
		return normalizeStringList([]string{typedValue})
	case []string:
		return normalizeStringList(typedValue)
	case []interface{}:
		items := make([]string, 0, len(typedValue))
		for _, item := range typedValue {
			items = append(items, cn03String(item))
		}
		return normalizeStringList(items)
	default:
		return normalizeStringList([]string{fmt.Sprintf("%v", typedValue)})
	}
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, rawValue := range values {
		for _, item := range strings.Split(rawValue, ",") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			if _, exists := seen[trimmed]; exists {
				continue
			}
			seen[trimmed] = struct{}{}
			normalized = append(normalized, trimmed)
		}
	}

	return normalized
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && trimmed != "<nil>" {
			return trimmed
		}
	}
	return ""
}

func cn03PeerOwnerID() string {
	for _, key := range []string{"CN03_PEER_OWNER_ID", "PEER_OWNER_ID"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}

func boolFromEvidence(value interface{}) bool {
	switch typedValue := value.(type) {
	case bool:
		return typedValue
	case string:
		return strings.EqualFold(strings.TrimSpace(typedValue), "true")
	default:
		return strings.EqualFold(strings.TrimSpace(fmt.Sprintf("%v", value)), "true")
	}
}
