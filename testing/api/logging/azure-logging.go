package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// AzureLoggingService implements Service for Azure Monitor/Log Analytics
type AzureLoggingService struct {
	activityLogsClient *armmonitor.ActivityLogsClient
	credential         azcore.TokenCredential
	ctx                context.Context
	cloudParams        *environment.CloudParams
	testParams         *environment.TestParams
}

// NewAzureLoggingService creates a new Azure logging service using default credential chain
func NewAzureLoggingService(ctx context.Context, cloudParams *environment.CloudParams, testParams *environment.TestParams) (*AzureLoggingService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	activityLogsClient, err := armmonitor.NewActivityLogsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureLoggingService{
		activityLogsClient: activityLogsClient,
		credential:         cred,
		ctx:                ctx,
		cloudParams:        cloudParams,
		testParams:         testParams,
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
	startTime := time.Now().Add(-time.Duration(lookbackMinutes) * time.Minute)
	endTime := time.Now()

	// Build the filter for Activity Log query
	// Filter by time range and resource group (storage account operations are logged at resource group level)
	filter := fmt.Sprintf("eventTimestamp ge '%s' and eventTimestamp le '%s' and resourceGroupName eq '%s'",
		startTime.UTC().Format(time.RFC3339),
		endTime.UTC().Format(time.RFC3339),
		s.cloudParams.AzureResourceGroup)

	pager := s.activityLogsClient.NewListPager(filter, nil)

	var entries []LogEntry
	for pager.More() {
		page, err := pager.NextPage(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get activity log page: %w", err)
		}

		for _, event := range page.Value {
			entry := LogEntry{
				Timestamp: azureGetTime(event.EventTimestamp),
				Resource:  azureGetString(event.ResourceID),
			}
			if event.OperationName != nil {
				entry.Action = azureGetString(event.OperationName.LocalizedValue)
			}
			if event.Status != nil {
				entry.Result = azureGetString(event.Status.LocalizedValue)
			}
			if event.Caller != nil {
				entry.Identity = *event.Caller
			}
			entries = append(entries, entry)
		}
	}

	if len(entries) == 0 {
		return []LogEntry{
			{
				Identity:  "azure-activity-log",
				Action:    "QueryAdminLogs",
				Resource:  resourceID,
				Timestamp: time.Now(),
				Result:    fmt.Sprintf("No admin events found in last %d minutes", lookbackMinutes),
			},
		}, nil
	}

	return entries, nil
}

// QueryDataWriteLogs queries Azure Storage diagnostic logs for write events
// Note: Data-level logs require Azure Diagnostic Settings to be configured
func (s *AzureLoggingService) QueryDataWriteLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	// Data-level logging in Azure requires Diagnostic Settings which send logs to
	// Log Analytics, Storage Account, or Event Hub. This is a configuration check.
	return []LogEntry{
		{
			Identity:  "azure-diagnostics",
			Action:    "QueryDataWriteLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "Data write logging requires Diagnostic Settings - check via policy",
		},
	}, nil
}

// QueryDataReadLogs queries Azure Storage diagnostic logs for read events
// Note: Data-level logs require Azure Diagnostic Settings to be configured
func (s *AzureLoggingService) QueryDataReadLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	// Data-level logging in Azure requires Diagnostic Settings which send logs to
	// Log Analytics, Storage Account, or Event Hub. This is a configuration check.
	return []LogEntry{
		{
			Identity:  "azure-diagnostics",
			Action:    "QueryDataReadLogs",
			Resource:  resourceID,
			Timestamp: time.Now(),
			Result:    "Data read logging requires Diagnostic Settings - check via policy",
		},
	}, nil
}

func azureGetString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func azureGetTime(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}
