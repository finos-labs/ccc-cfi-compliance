package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/finos-labs/ccc-cfi-compliance/testing/types"
)

// AzureLoggingService implements Service for Azure Monitor/Log Analytics
type AzureLoggingService struct {
	activityLogsClient *armmonitor.ActivityLogsClient
	logsClient         *azquery.LogsClient
	credential         azcore.TokenCredential
	ctx                context.Context
	cloudParams *types.CloudParams
	instance    types.InstanceConfig
	testParams         *types.TestParams
}

// NewAzureLoggingService creates a new Azure logging service using default credential chain
func NewAzureLoggingService(ctx context.Context, cloudParams *types.CloudParams, testParams *types.TestParams, instance types.InstanceConfig) (*AzureLoggingService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	activityLogsClient, err := armmonitor.NewActivityLogsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	logsClient, err := azquery.NewLogsClient(cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureLoggingService{
		activityLogsClient: activityLogsClient,
		logsClient:         logsClient,
		credential:         cred,
		ctx:                ctx,
		cloudParams: cloudParams,
		instance:    instance,
		testParams:         testParams,
	}, nil
}

// TestParams returns the test parameters
func (s *AzureLoggingService) TestParams() *types.TestParams {
	return s.testParams
}

// CloudParams returns the cloud-specific parameters
func (s *AzureLoggingService) CloudParams() *types.CloudParams {
	return s.cloudParams
}

// GetOrProvisionTestableResources returns testable resources for the logging service
func (s *AzureLoggingService) GetOrProvisionTestableResources() ([]types.TestParams, error) {
	resourceName := "azure-monitor"
	return []types.TestParams{
		{
			ServiceType:         "logging",
			ProviderServiceType: "azure-monitor",
			CatalogTypes:        []string{"CCC.Core"},
			TagFilter:           []string{"@logging", "@PerService"},
			ResourceName:        resourceName,
			UID:                 resourceName,
			ReportFile:          "azure-monitor",
			ReportTitle:         "Azure Monitor",
			Instance:   s.instance,
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

	return entries, nil
}

// QueryDataWriteLogs queries Azure Log Analytics for storage write events
// Note: Requires Diagnostic Settings configured to send StorageWrite logs to a Log Analytics workspace
func (s *AzureLoggingService) QueryDataWriteLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return s.queryStorageLogs(resourceID, lookbackMinutes, "StorageWrite")
}

// QueryDataReadLogs queries Azure Log Analytics for storage read events
// Note: Requires Diagnostic Settings configured to send StorageRead logs to a Log Analytics workspace
func (s *AzureLoggingService) QueryDataReadLogs(resourceID string, lookbackMinutes int) ([]LogEntry, error) {
	return s.queryStorageLogs(resourceID, lookbackMinutes, "StorageRead")
}

func (s *AzureLoggingService) queryStorageLogs(resourceID string, lookbackMinutes int, category string) ([]LogEntry, error) {
	workspaceID := serviceParamString(s.instance.ServiceProperties("logging"), "azure-log-analytics-workspace-id")
	if workspaceID == "" {
		return []LogEntry{}, nil
	}

	storageAccount := serviceParamString(s.instance.ServiceProperties("object-storage"), "azure-storage-account")

	kql := fmt.Sprintf(`StorageBlobLogs
| where TimeGenerated >= ago(%dm)
| where Category == '%s'
| where AccountName == '%s'
| project TimeGenerated, CallerIpAddress, AuthenticationType, OperationName, StatusText, Uri
| order by TimeGenerated desc`,
		lookbackMinutes, category, storageAccount)

	query := azquery.Body{Query: &kql}
	resp, err := s.logsClient.QueryWorkspace(s.ctx, workspaceID, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query Log Analytics workspace: %w", err)
	}

	var entries []LogEntry
	for _, table := range resp.Tables {
		colIdx := map[string]int{}
		for i, col := range table.Columns {
			if col.Name != nil {
				colIdx[*col.Name] = i
			}
		}
		for _, row := range table.Rows {
			entry := LogEntry{}
			if i, ok := colIdx["CallerIpAddress"]; ok && i < len(row) && row[i] != nil {
				entry.Identity = fmt.Sprintf("%v", row[i])
			}
			if i, ok := colIdx["OperationName"]; ok && i < len(row) && row[i] != nil {
				entry.Action = fmt.Sprintf("%v", row[i])
			}
			if i, ok := colIdx["TimeGenerated"]; ok && i < len(row) && row[i] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", row[i])); err == nil {
					entry.Timestamp = t
				}
			}
			if i, ok := colIdx["Uri"]; ok && i < len(row) && row[i] != nil {
				entry.Resource = fmt.Sprintf("%v", row[i])
			}
			if i, ok := colIdx["StatusText"]; ok && i < len(row) && row[i] != nil {
				entry.Result = fmt.Sprintf("%v", row[i])
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
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
