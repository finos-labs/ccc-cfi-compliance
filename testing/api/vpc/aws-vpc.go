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
	// VPC controls are evaluated from control-plane inventory in a region.
	// Represent the region/account as a single testable "resource".
	return []environment.TestParams{
		{
			ResourceName:        fmt.Sprintf("aws-vpc-%s", s.cloudParams.Region),
			UID:                 fmt.Sprintf("aws://vpc/%s", s.cloudParams.Region),
			ProviderServiceType: "ec2:vpc",
			ServiceType:         "vpc",
			CatalogTypes:        []string{"CCC.VPC"},
			CloudParams:         s.cloudParams,
		},
	}, nil
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

