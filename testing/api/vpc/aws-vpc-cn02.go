package vpc

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

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
