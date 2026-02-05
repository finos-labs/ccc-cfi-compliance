package vpc

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

func tagValue(tags []types.Tag, key string) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == key {
			return aws.ToString(t.Value)
		}
	}
	return ""
}
