package vpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	ccctypes "github.com/finos-labs/ccc-cfi-compliance/testing/types"
)

// AWSVPCService implements VPC Service for AWS EC2/VPC.
type AWSVPCService struct {
	client      *ec2.Client
	ctx         context.Context
	instance ccctypes.InstanceConfig
}

// NewAWSVPCService creates a new AWS VPC service using default credentials.
func NewAWSVPCService(ctx context.Context, instance ccctypes.InstanceConfig) (*AWSVPCService, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(instance.Properties.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSVPCService{
		client:   ec2.NewFromConfig(cfg),
		ctx:      ctx,
		instance: instance,
	}, nil
}

func (s *AWSVPCService) GetOrProvisionTestableResources() ([]ccctypes.TestParams, error) {
	// Return all VPCs in the configured region as testable resources.
	// Some controls are region-scoped, but returning per-VPC resources allows
	// controls that require a VPC ID (e.g., subnet-level checks) to execute.
	output, err := s.client.DescribeVpcs(s.ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	resources := make([]ccctypes.TestParams, 0, len(output.Vpcs))
	for _, vpc := range output.Vpcs {
		vpcID := aws.ToString(vpc.VpcId)
		resourceName := vpcID
		if nameTag := tagValue(vpc.Tags, "Name"); nameTag != "" {
			resourceName = nameTag
		}

		resources = append(resources, ccctypes.TestParams{
			ResourceName:        resourceName,
			UID:                 vpcID,
			ProviderServiceType: "ec2:vpc",
			ServiceType:         "vpc",
			CatalogTypes:        []string{"CCC.VPC"},
			TagFilter:           []string{"@vpc", "@CCC.VPC"},
			Instance:            s.instance,
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
func (s *AWSVPCService) UpdateResourcePolicy() error       { return nil }
func (s *AWSVPCService) TriggerDataWrite(_ string) error   { return nil }
func (s *AWSVPCService) TearDown() error                   { return nil }
func (s *AWSVPCService) GetResourceRegion(_ string) (string, error) {
	return s.instance.Properties.Region, nil
}
func (s *AWSVPCService) GetReplicationStatus(_ string) (*generic.ReplicationStatus, error) {
	return nil, fmt.Errorf("replication status not applicable for VPC service")
}

func tagValue(tags []types.Tag, key string) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == key {
			return aws.ToString(t.Value)
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
