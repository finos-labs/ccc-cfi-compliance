package vpc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

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

	publicSubnets, err := s.listPublicSubnets(vpcIDStr)
	if err != nil {
		return "", err
	}

	if len(publicSubnets) == 0 {
		return fmt.Sprintf("CCC.VPC.CN02.AR01: N/A (no public subnets found for VPC %s)", vpcIDStr), nil
	}

	return fmt.Sprintf("CCC.VPC.CN02.AR01: checking %d public subnet(s) for VPC %s", len(publicSubnets), vpcIDStr), nil
}

func (s *AWSVPCService) ListVpcFlowLogs(vpcID string) ([]interface{}, error) {
	return s.listVpcFlowLogs(vpcID)
}

func (s *AWSVPCService) HasActiveAllTrafficFlowLogs(vpcID string) (bool, error) {
	flowLogs, err := s.listVpcFlowLogs(vpcID)
	if err != nil {
		return false, err
	}
	if len(flowLogs) == 0 {
		return false, nil
	}

	for _, item := range flowLogs {
		row, ok := item.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unexpected flow log record type")
		}

		status := fmt.Sprintf("%v", row["FlowLogStatus"])
		trafficType := fmt.Sprintf("%v", row["TrafficType"])
		if status != "ACTIVE" || trafficType != "ALL" {
			return false, nil
		}
	}

	return true, nil
}

func (s *AWSVPCService) SummarizeVpcFlowLogs(vpcID string) (string, error) {
	vpcIDStr := fmt.Sprintf("%v", vpcID)
	if vpcIDStr == "" {
		return "", fmt.Errorf("vpcID is required")
	}

	flowLogs, err := s.listVpcFlowLogs(vpcIDStr)
	if err != nil {
		return "", err
	}

	if len(flowLogs) == 0 {
		return fmt.Sprintf("CCC.VPC.CN04.AR01: N/A (no VPC flow logs configured for VPC %s)", vpcIDStr), nil
	}

	return fmt.Sprintf("CCC.VPC.CN04.AR01: checking %d flow log record(s) for VPC %s", len(flowLogs), vpcIDStr), nil
}

func (s *AWSVPCService) AttemptDisallowedPeeringDryRun(requesterVpcID string) (map[string]interface{}, error) {
	requesterID := fmt.Sprintf("%v", requesterVpcID)
	if requesterID == "" {
		return nil, fmt.Errorf("requesterVpcID is required")
	}

	peerVpcID, peerOwnerID := cn03PeerInputs()
	if peerVpcID == "" {
		return nil, fmt.Errorf("missing required peer VPC input: set PEER_VPC_ID")
	}

	evidence := map[string]interface{}{
		"RequesterVpcId": requesterID,
		"PeerVpcId":      peerVpcID,
		"PeerOwnerId":    peerOwnerID,
		"ExitCode":       0,
		"DryRunAllowed":  false,
		"ErrorCode":      "",
		"ErrorMessage":   "",
		"Stderr":         "",
	}

	in := &ec2.CreateVpcPeeringConnectionInput{
		VpcId:     aws.String(requesterID),
		PeerVpcId: aws.String(peerVpcID),
		DryRun:    aws.Bool(true),
	}
	if peerOwnerID != "" {
		in.PeerOwnerId = aws.String(peerOwnerID)
	}

	_, err := s.client.CreateVpcPeeringConnection(s.ctx, in)
	if err == nil {
		// Defensive fallback: dry-run should return an error, but if not,
		// treat it as allowed because the action path was not denied.
		evidence["DryRunAllowed"] = true
		return evidence, nil
	}

	evidence["ExitCode"] = 1
	evidence["Stderr"] = err.Error()

	var apiErr smithy.APIError
	if ok := errors.As(err, &apiErr); ok {
		code := apiErr.ErrorCode()
		msg := apiErr.ErrorMessage()
		evidence["ErrorCode"] = code
		evidence["ErrorMessage"] = msg
		if code == "DryRunOperation" {
			evidence["DryRunAllowed"] = true
		}
	} else if strings.Contains(err.Error(), "DryRunOperation") {
		evidence["DryRunAllowed"] = true
	}

	return evidence, nil
}

func (s *AWSVPCService) IsDisallowedPeeringPrevented(requesterVpcID string) (bool, error) {
	evidence, err := s.AttemptDisallowedPeeringDryRun(requesterVpcID)
	if err != nil {
		return false, err
	}

	return s.EvaluateDisallowedPeeringDryRun(evidence)
}

func (s *AWSVPCService) evaluatePeeringOutcomeForMode(evidence map[string]interface{}, mode string) (bool, error) {
	switch mode {
	case "disallowed":
		return s.EvaluateDisallowedPeeringDryRun(evidence)
	case "allowed":
		return s.EvaluateAllowedPeeringDryRun(evidence)
	default:
		return false, fmt.Errorf("unsupported CN03 mode: %s", mode)
	}
}

func (s *AWSVPCService) EvaluateDisallowedPeeringDryRun(evidence map[string]interface{}) (bool, error) {
	if evidence == nil {
		return false, fmt.Errorf("evidence is required")
	}

	if dryRunAllowed, ok := evidence["DryRunAllowed"].(bool); ok && dryRunAllowed {
		return false, nil
	}

	errorCode := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", evidence["ErrorCode"])))
	if errorCode != "" {
		switch errorCode {
		case "accessdenied", "unauthorizedoperation", "authfailure":
			return true, nil
		case "dryrunoperation":
			return false, nil
		}
	}

	stderrLower := strings.ToLower(fmt.Sprintf("%v", evidence["Stderr"]))
	for _, marker := range []string{
		"accessdenied",
		"unauthorizedoperation",
		"authfailure",
		"not authorized",
		"explicit deny",
		"denied",
	} {
		if strings.Contains(stderrLower, marker) {
			return true, nil
		}
	}

	return false, nil
}

func (s *AWSVPCService) EvaluateAllowedPeeringDryRun(evidence map[string]interface{}) (bool, error) {
	if evidence == nil {
		return false, fmt.Errorf("evidence is required")
	}

	allowedListReference := cn03AllowedListReference()
	if allowedListReference == "" {
		return false, fmt.Errorf("allowed-list basis is undefined: set CN03_ALLOWED_LIST_REFERENCE for allowed-mode validation")
	}

	if dryRunAllowed, ok := evidence["DryRunAllowed"].(bool); ok && dryRunAllowed {
		return true, nil
	}

	errorCode := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", evidence["ErrorCode"])))
	return errorCode == "dryrunoperation", nil
}

func (s *AWSVPCService) SummarizePeeringOutcomeCompact(evidence map[string]interface{}, mode string) (map[string]interface{}, error) {
	if evidence == nil {
		return nil, fmt.Errorf("evidence is required")
	}

	normalizedMode := canonicalizeCN03Mode(mode)
	if normalizedMode == "" {
		normalizedMode = "disallowed"
	}

	passed, err := s.evaluatePeeringOutcomeForMode(evidence, normalizedMode)
	verdict := "FAIL"
	resultClass := "FAIL"
	if err == nil && passed {
		verdict = "PASS"
		resultClass = "PASS"
	} else if err != nil {
		resultClass = "SETUP_ERROR"
	}

	dryRunAllowed := boolFromEvidence(evidence["DryRunAllowed"])
	errorCode := strings.TrimSpace(fmt.Sprintf("%v", evidence["ErrorCode"]))
	allowedListRef := cn03AllowedListReference()
	reason := compactPeeringReason(normalizedMode, dryRunAllowed, errorCode, passed, err)

	return map[string]interface{}{
		"ControlId":         "CCC.VPC.CN03.AR01",
		"Mode":              normalizedMode,
		"Verdict":           verdict,
		"ResultClass":       resultClass,
		"Reason":            reason,
		"RequesterVpcId":    strings.TrimSpace(fmt.Sprintf("%v", evidence["RequesterVpcId"])),
		"PeerVpcId":         strings.TrimSpace(fmt.Sprintf("%v", evidence["PeerVpcId"])),
		"PeerOwnerId":       strings.TrimSpace(fmt.Sprintf("%v", evidence["PeerOwnerId"])),
		"DryRunAllowed":     dryRunAllowed,
		"ErrorCode":         errorCode,
		"AllowedListBasis":  allowedListRefOrDefault(allowedListRef),
		"ExpectedCondition": expectedConditionForMode(normalizedMode),
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

func tagValue(tags []types.Tag, key string) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == key {
			return aws.ToString(t.Value)
		}
	}
	return ""
}

func cn03PeerInputs() (peerVpcID string, peerOwnerID string) {
	peerVpcID = strings.TrimSpace(os.Getenv("PEER_VPC_ID"))
	if peerVpcID == "" {
		peerVpcID = strings.TrimSpace(os.Getenv("CN03_PEER_VPC_ID"))
	}

	peerOwnerID = strings.TrimSpace(os.Getenv("PEER_OWNER_ID"))
	if peerOwnerID == "" {
		peerOwnerID = strings.TrimSpace(os.Getenv("CN03_PEER_OWNER_ID"))
	}

	return peerVpcID, peerOwnerID
}

func canonicalizeCN03Mode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))

	switch mode {
	case "disallowed", "denied", "blocked":
		return "disallowed"
	case "allowed", "permitted":
		return "allowed"
	default:
		return mode
	}
}

func cn03AllowedListReference() string {
	return strings.TrimSpace(os.Getenv("CN03_ALLOWED_LIST_REFERENCE"))
}

func allowedListRefOrDefault(value string) string {
	if value == "" {
		return "not-required"
	}
	return value
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

func expectedConditionForMode(mode string) string {
	switch mode {
	case "allowed":
		return "DryRunAllowed=true (DryRunOperation) with allowlist basis"
	default:
		return "DryRunAllowed=false with deny indicator"
	}
}

func compactPeeringReason(mode string, dryRunAllowed bool, errorCode string, passed bool, evalErr error) string {
	if evalErr != nil {
		return evalErr.Error()
	}

	if mode == "allowed" {
		if passed {
			return "target treated as explicitly allowed; dry-run indicates request would succeed"
		}
		return "target not proven as allowed or request denied"
	}

	if passed {
		return "target treated as disallowed; dry-run denied as expected"
	}

	if dryRunAllowed || strings.EqualFold(errorCode, "DryRunOperation") {
		return "request would be allowed (DryRunOperation)"
	}
	return "deny indicator was not detected"
}
