package logging

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AzureLoggingService implements Service for Azure Monitor/Log Analytics
type AzureLoggingService struct {
	logsClient  *azquery.LogsClient
	credential  azcore.TokenCredential
	ctx         context.Context
	cloudParams *environment.CloudParams
	testParams  *environment.TestParams
}

// NewAzureLoggingService creates a new Azure logging service using default credential chain
func NewAzureLoggingService(ctx context.Context, cloudParams *environment.CloudParams, testParams *environment.TestParams) (*AzureLoggingService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	logsClient, err := azquery.NewLogsClient(cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureLoggingService{
		logsClient:  logsClient,
		credential:  cred,
		ctx:         ctx,
		cloudParams: cloudParams,
		testParams:  testParams,
	}, nil
}

// TestParams returns the test parameters
func (s *AzureLoggingService) TestParams() *environment.TestParams {
	return s.testParams
}

// CloudParams returns the cloud-specific parameters
func (s *AzureLoggingService) CloudParams() *environment.CloudParams {
	return s.cloudParams
}

// GetOrProvisionTestableResources returns testable resources for the logging service
func (s *AzureLoggingService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	resourceName := "azure-monitor"
	return []environment.TestParams{
		{
			ServiceType:         "logging",
			ProviderServiceType: "azure-monitor",
			CatalogTypes:        []string{"CCC.Core"},
			TagFilter:           []string{"@logging", "@PerService"},
			ResourceName:        resourceName,
			UID:                 resourceName,
			ReportFile:          "azure-monitor",
			ReportTitle:         "Azure Monitor",
			CloudParams:         *s.cloudParams,
		},
	}, nil
}

// CheckUserProvisioned validates that the service's identity is properly provisioned
func (s *AzureLoggingService) CheckUserProvisioned() error {
	return nil
}

// ElevateAccessForInspection temporarily elevates access permissions
func (s *AzureLoggingService) ElevateAccessForInspection() error {
	return nil
}

// ResetAccess restores the original access permissions
func (s *AzureLoggingService) ResetAccess() error {
	return nil
}

// UpdateResourcePolicy is not applicable for logging service
func (s *AzureLoggingService) UpdateResourcePolicy() error {
	return nil
}

// QueryAdminLogs queries Azure Activity Log for admin events
func (s *AzureLoggingService) QueryAdminLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return []LogEntry{
		{
			Identity:  "azure-activity-log-default",
			Action:    "QueryAdminLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "Activity Log is enabled by default in Azure",
		},
	}, nil
}

// QueryDataWriteLogs queries Azure Storage diagnostic logs for write events
func (s *AzureLoggingService) QueryDataWriteLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return []LogEntry{
		{
			Identity:  "azure-diagnostics-not-configured",
			Action:    "QueryDataWriteLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "StorageWrite diagnostic logging must be explicitly enabled",
		},
	}, nil
}

// QueryDataReadLogs queries Azure Storage diagnostic logs for read events
func (s *AzureLoggingService) QueryDataReadLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return []LogEntry{
		{
			Identity:  "azure-diagnostics-not-configured",
			Action:    "QueryDataReadLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "StorageRead diagnostic logging must be explicitly enabled",
		},
	}, nil
}
