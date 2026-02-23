package logging

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AWSLoggingService implements Service for AWS CloudTrail
type AWSLoggingService struct {
	*generic.AWSService
	cloudTrailClient *cloudtrail.Client
	ctx              context.Context
	cloudParams      *environment.CloudParams
	testParams       *environment.TestParams
}

// NewAWSLoggingService creates a new AWS logging service using default credential chain
func NewAWSLoggingService(ctx context.Context, cloudParams *environment.CloudParams, testParams *environment.TestParams) (*AWSLoggingService, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cloudParams.Region))
	if err != nil {
		return nil, err
	}

	return &AWSLoggingService{
		AWSService:       generic.NewAWSService(ctx),
		cloudTrailClient: cloudtrail.NewFromConfig(cfg),
		ctx:              ctx,
		cloudParams:      cloudParams,
		testParams:       testParams,
	}, nil
}

// NewAWSLoggingServiceWithCredentials creates a new AWS logging service with explicit credentials
func NewAWSLoggingServiceWithCredentials(ctx context.Context, cloudParams *environment.CloudParams, testParams *environment.TestParams, accessKeyID, secretAccessKey, sessionToken string) (*AWSLoggingService, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cloudParams.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken)),
	)
	if err != nil {
		return nil, err
	}

	return &AWSLoggingService{
		AWSService:       generic.NewAWSService(ctx),
		cloudTrailClient: cloudtrail.NewFromConfig(cfg),
		ctx:              ctx,
		cloudParams:      cloudParams,
		testParams:       testParams,
	}, nil
}

// TestParams returns the test parameters
func (s *AWSLoggingService) TestParams() *environment.TestParams {
	return s.testParams
}

// CloudParams returns the cloud-specific parameters
func (s *AWSLoggingService) CloudParams() *environment.CloudParams {
	return s.cloudParams
}

// GetOrProvisionTestableResources returns testable resources for the logging service
func (s *AWSLoggingService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	trailName := s.DiscoverCloudTrailName()

	return []environment.TestParams{
		{
			ServiceType:         "logging",
			ProviderServiceType: "cloudtrail",
			CatalogTypes:        []string{"CCC.Core"},
			TagFilter:           []string{"@logging", "@PerService"},
			ResourceName:        trailName,
			UID:                 trailName,
			ReportFile:          "cloudtrail-" + trailName,
			ReportTitle:         "CloudTrail: " + trailName,
			CloudParams:         *s.cloudParams,
		},
	}, nil
}

// CheckUserProvisioned validates that the service's identity is properly provisioned
func (s *AWSLoggingService) CheckUserProvisioned() error {
	return nil
}

// ElevateAccessForInspection temporarily elevates access permissions
func (s *AWSLoggingService) ElevateAccessForInspection() error {
	return nil
}

// ResetAccess restores the original access permissions
func (s *AWSLoggingService) ResetAccess() error {
	return nil
}

// UpdateResourcePolicy is not applicable for logging service
func (s *AWSLoggingService) UpdateResourcePolicy() error {
	return nil
}

// QueryAdminLogs queries CloudTrail for admin/management events
func (s *AWSLoggingService) QueryAdminLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return s.queryCloudTrailLogs(resourceID, lookbackMinutes, "management")
}

// QueryDataWriteLogs queries CloudTrail for data write events
func (s *AWSLoggingService) QueryDataWriteLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return s.queryCloudTrailLogs(resourceID, lookbackMinutes, "data-write")
}

// QueryDataReadLogs queries CloudTrail for data read events
func (s *AWSLoggingService) QueryDataReadLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return s.queryCloudTrailLogs(resourceID, lookbackMinutes, "data-read")
}

func (s *AWSLoggingService) queryCloudTrailLogs(resourceID string, lookbackMinutes int, eventType string) ([]LogEntry, error) {
	startTime := time.Now().Add(-time.Duration(lookbackMinutes) * time.Minute)
	endTime := time.Now()

	input := &cloudtrail.LookupEventsInput{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	result, err := s.cloudTrailClient.LookupEvents(s.ctx, input)
	if err != nil {
		return nil, err
	}

	var entries []LogEntry
	for _, event := range result.Events {
		entry := LogEntry{
			Timestamp: *event.EventTime,
			Action:    getString(event.EventName),
			Resource:  getString(event.EventSource),
		}
		if event.Username != nil {
			entry.Identity = *event.Username
		}
		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return []LogEntry{
			{
				Identity:  "cloudtrail",
				Action:    "QueryLogs",
				Resource:  resourceID,
				Timestamp: time.Now(),
				Result:    "No " + eventType + " events found in lookback period",
			},
		}, nil
	}

	return entries, nil
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
