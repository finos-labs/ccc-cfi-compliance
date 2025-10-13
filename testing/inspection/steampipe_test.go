package inspection

import (
	"context"
	"testing"
	"time"
)

func TestGetAccessiblePorts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	providers := []string{"aws", "azure", "gcp"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			ports, err := GetAccessiblePorts(ctx, provider)

			// If Steampipe isn't running, skip the test
			if err != nil && containsString(err.Error(), "failed to connect to Steampipe") {
				t.Skipf("Steampipe not running, skipping test for %s", provider)
				return
			}

			if err != nil && containsString(err.Error(), "failed to ping Steampipe") {
				t.Skipf("Steampipe not accessible, skipping test for %s", provider)
				return
			}

			// If permission error (expected for GCP without proper setup), skip
			if err != nil && (containsString(err.Error(), "Error 403") || containsString(err.Error(), "Permission") || containsString(err.Error(), "forbidden")) {
				t.Skipf("Permission denied for %s (credentials not configured), skipping test", provider)
				return
			}

			if err != nil {
				t.Fatalf("GetAccessiblePorts(%s) returned error: %v", provider, err)
			}

			t.Logf("Found %d accessible ports for provider %s", len(ports), provider)

			// Validate structure of returned data
			for i, port := range ports {
				// Check that required fields are populated
				if port.Provider != provider {
					t.Errorf("Port[%d].Provider = %q, want %q", i, port.Provider, provider)
				}
				if port.UID == "" {
					t.Errorf("Port[%d].UID is empty", i)
				}
				if port.Region == "" {
					t.Errorf("Port[%d].Region is empty", i)
				}

				// Log all ports with detailed information
				t.Logf("Port[%d]: UID=%s, Port=%s, Protocol=%s, ProviderServiceType=%s, CatalogType=%s, Region=%s, HostName=%s",
					i, port.UID, port.PortNumber, port.Protocol, port.ProviderServiceType, port.CatalogType, port.Region, port.HostName)
			}
		})
	}
}

func TestGetServices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	providers := []string{"aws", "azure", "gcp"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			services, err := GetServices(ctx, provider)

			// If Steampipe isn't running, skip the test
			if err != nil && containsString(err.Error(), "failed to connect to Steampipe") {
				t.Skipf("Steampipe not running, skipping test for %s", provider)
				return
			}

			if err != nil && containsString(err.Error(), "failed to ping Steampipe") {
				t.Skipf("Steampipe not accessible, skipping test for %s", provider)
				return
			}

			// If permission error (expected for GCP without proper setup), skip
			if err != nil && (containsString(err.Error(), "Error 403") || containsString(err.Error(), "Permission") || containsString(err.Error(), "forbidden")) {
				t.Skipf("Permission denied for %s (credentials not configured), skipping test", provider)
				return
			}

			if err != nil {
				t.Fatalf("GetServices(%s) returned error: %v", provider, err)
			}

			t.Logf("Found %d services for provider %s", len(services), provider)

			// Validate structure of returned data
			for i, svc := range services {
				// Check that required fields are populated
				if svc.Provider != provider {
					t.Errorf("Service[%d].Provider = %q, want %q", i, svc.Provider, provider)
				}
				if svc.UID == "" {
					t.Errorf("Service[%d].UID is empty", i)
				}
				if svc.Region == "" {
					t.Errorf("Service[%d].Region is empty", i)
				}
				if svc.ServiceType == "" {
					t.Errorf("Service[%d].ServiceType is empty", i)
				}

				// PortNumber should be empty for services
				if svc.PortNumber != "" {
					t.Logf("Note: Service[%d] has PortNumber=%s (expected to be empty)", i, svc.PortNumber)
				}

				// Log all services with detailed information
				t.Logf("Service[%d]: UID=%s, ProviderServiceType=%s, CatalogType=%s, HostName=%s, Region=%s, Provider=%s, Labels=%v",
					i, svc.UID, svc.ProviderServiceType, svc.CatalogType, svc.HostName, svc.Region, svc.Provider, svc.Labels)
			}
		})
	}
}

func TestGetAccessiblePortsInvalidProvider(t *testing.T) {
	ctx := context.Background()

	_, err := GetAccessiblePorts(ctx, "invalid")
	if err == nil {
		t.Fatal("GetAccessiblePorts with invalid provider should return error")
	}

	if !containsString(err.Error(), "unsupported provider") {
		t.Errorf("Expected 'unsupported provider' error, got: %v", err)
	}
}

func TestGetServicesInvalidProvider(t *testing.T) {
	ctx := context.Background()

	_, err := GetServices(ctx, "invalid")
	if err == nil {
		t.Fatal("GetServices with invalid provider should return error")
	}

	if !containsString(err.Error(), "unsupported provider") {
		t.Errorf("Expected 'unsupported provider' error, got: %v", err)
	}
}

func TestConnectSteampipe(t *testing.T) {
	db, err := connectSteampipe()

	// If Steampipe isn't running, skip the test
	if err != nil {
		t.Skipf("Steampipe not accessible: %v", err)
		return
	}

	defer db.Close()

	// Try a simple query
	rows, err := db.Query("SELECT 1")
	if err != nil {
		t.Fatalf("Failed to execute test query: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Expected at least one row from test query")
	}

	var result int
	if err := rows.Scan(&result); err != nil {
		t.Fatalf("Failed to scan result: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected result = 1, got %d", result)
	}

	t.Log("Successfully connected to Steampipe")
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
