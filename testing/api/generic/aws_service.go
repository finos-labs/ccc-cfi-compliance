package generic

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
)

// AWSService provides common AWS functionality that can be embedded in service implementations
type AWSService struct {
	ctx             context.Context
	cloudTrailName  string
	cloudTrailCache bool
}

// NewAWSService creates a new AWSService with the given context
func NewAWSService(ctx context.Context) *AWSService {
	return &AWSService{
		ctx: ctx,
	}
}

// DiscoverCloudTrailName finds the CloudTrail trail name for the account
// Priority:
// 1. AWS_CLOUDTRAIL_NAME environment variable (allows override)
// 2. First multi-region trail found via CloudTrail API
// 3. First trail found via CloudTrail API
// Returns empty string if no trail is found
func (s *AWSService) DiscoverCloudTrailName() string {
	// Return cached value if already discovered
	if s.cloudTrailCache {
		return s.cloudTrailName
	}
	s.cloudTrailCache = true

	// Check environment variable first (allows user override)
	if trailName := os.Getenv("AWS_CLOUDTRAIL_NAME"); trailName != "" {
		s.cloudTrailName = trailName
		return s.cloudTrailName
	}

	// Query CloudTrail API
	cfg, err := config.LoadDefaultConfig(s.ctx)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load AWS config for CloudTrail discovery: %v\n", err)
		return ""
	}

	client := cloudtrail.NewFromConfig(cfg)
	result, err := client.DescribeTrails(s.ctx, &cloudtrail.DescribeTrailsInput{})
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to describe CloudTrail trails: %v\n", err)
		return ""
	}

	if len(result.TrailList) == 0 {
		fmt.Printf("⚠️  Warning: No CloudTrail trails found in account\n")
		return ""
	}

	// Prefer multi-region trails as they capture all events
	for _, trail := range result.TrailList {
		if trail.IsMultiRegionTrail != nil && *trail.IsMultiRegionTrail && trail.Name != nil {
			s.cloudTrailName = *trail.Name
			fmt.Printf("✅ Discovered multi-region CloudTrail: %s\n", s.cloudTrailName)
			return s.cloudTrailName
		}
	}

	// Fall back to first trail
	if result.TrailList[0].Name != nil {
		s.cloudTrailName = *result.TrailList[0].Name
		fmt.Printf("✅ Discovered CloudTrail: %s\n", s.cloudTrailName)
	}

	return s.cloudTrailName
}

// GetCloudTrailName returns the cached CloudTrail name or discovers it
func (s *AWSService) GetCloudTrailName() string {
	if !s.cloudTrailCache {
		return s.DiscoverCloudTrailName()
	}
	return s.cloudTrailName
}
