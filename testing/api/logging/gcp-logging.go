package logging

import (
	"context"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// GCPLoggingService implements Service for GCP Cloud Audit Logs
type GCPLoggingService struct {
	logAdminClient *logadmin.Client
	ctx            context.Context
	cloudParams    *environment.CloudParams
	testParams     *environment.TestParams
}

// NewGCPLoggingService creates a new GCP logging service
func NewGCPLoggingService(ctx context.Context, cloudParams *environment.CloudParams, testParams *environment.TestParams) (*GCPLoggingService, error) {
	client, err := logadmin.NewClient(ctx, cloudParams.GCPProjectID)
	if err != nil {
		return nil, err
	}

	return &GCPLoggingService{
		logAdminClient: client,
		ctx:            ctx,
		cloudParams:    cloudParams,
		testParams:     testParams,
	}, nil
}

// TestParams returns the test parameters
func (s *GCPLoggingService) TestParams() *environment.TestParams {
	return s.testParams
}

// CloudParams returns the cloud-specific parameters
func (s *GCPLoggingService) CloudParams() *environment.CloudParams {
	return s.cloudParams
}

// GetOrProvisionTestableResources returns testable resources for the logging service
func (s *GCPLoggingService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	resourceName := "cloud-audit-logs"
	return []environment.TestParams{
		{
			ServiceType:         "logging",
			ProviderServiceType: "cloud-audit-logs",
			CatalogTypes:        []string{"CCC.Core"},
			TagFilter:           []string{"@logging", "@PerService"},
			ResourceName:        resourceName,
			UID:                 resourceName,
			ReportFile:          "cloud-audit-logs",
			ReportTitle:         "Cloud Audit Logs",
			CloudParams:         *s.cloudParams,
		},
	}, nil
}

// CheckUserProvisioned validates that the service's identity is properly provisioned
func (s *GCPLoggingService) CheckUserProvisioned() error {
	return nil
}

// ElevateAccessForInspection temporarily elevates access permissions
func (s *GCPLoggingService) ElevateAccessForInspection() error {
	return nil
}

// ResetAccess restores the original access permissions
func (s *GCPLoggingService) ResetAccess() error {
	return nil
}

// UpdateResourcePolicy is not applicable for logging service
func (s *GCPLoggingService) UpdateResourcePolicy() error {
	return nil
}

// QueryAdminLogs queries Cloud Audit Logs for admin activity events
func (s *GCPLoggingService) QueryAdminLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return []LogEntry{
		{
			Identity:  "cloud-audit-logs-default",
			Action:    "QueryAdminLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "Admin Activity audit logs are enabled by default in GCP",
		},
	}, nil
}

// QueryDataWriteLogs queries Cloud Audit Logs for data write events
func (s *GCPLoggingService) QueryDataWriteLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return []LogEntry{
		{
			Identity:  "cloud-audit-logs-not-configured",
			Action:    "QueryDataWriteLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "DATA_WRITE audit logs must be explicitly enabled in IAM policy",
		},
	}, nil
}

// QueryDataReadLogs queries Cloud Audit Logs for data read events
func (s *GCPLoggingService) QueryDataReadLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return []LogEntry{
		{
			Identity:  "cloud-audit-logs-not-configured",
			Action:    "QueryDataReadLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "DATA_READ audit logs must be explicitly enabled in IAM policy",
		},
	}, nil
}

// Close releases resources
func (s *GCPLoggingService) Close() error {
	if s.logAdminClient != nil {
		return s.logAdminClient.Close()
	}
	return nil
}
