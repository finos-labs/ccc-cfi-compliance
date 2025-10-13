package runner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
)

const (
	// featuresPath is hardcoded relative to this file (testing/runner/main_test.go)
	// Features are in testing/features/
	featuresPath = "../features"
)

var (
	provider     = flag.String("provider", "", "Cloud provider (aws, azure, or gcp)")
	outputDir    = flag.String("output", "output", "Output directory for test reports")
	timeout      = flag.Duration("timeout", 30*time.Minute, "Timeout for all tests")
	skipPorts    = flag.Bool("skip-ports", false, "Skip port tests")
	skipServices = flag.Bool("skip-services", false, "Skip service tests")
)

func TestRunCompliance(t *testing.T) {

	// Validate required flags
	if *provider == "" {
		log.Fatal("Error: -provider flag is required (aws, azure, or gcp)")
	}

	if *provider != "aws" && *provider != "azure" && *provider != "gcp" {
		log.Fatalf("Error: invalid provider '%s' (must be aws, azure, or gcp)", *provider)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	log.Printf("🚀 Starting CCC CFI Compliance Tests")
	log.Printf("   Provider: %s", *provider)
	log.Printf("   Output Directory: %s", *outputDir)
	log.Printf("   Features Path: %s", featuresPath)
	log.Printf("   Timeout: %s", *timeout)
	log.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedTests := 0

	// Run port tests
	if !*skipPorts {
		log.Println("🔍 Discovering accessible ports...")
		ports, err := inspection.GetAccessiblePorts(ctx, *provider)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to discover ports: %v", err)
		} else {
			log.Printf("   Found %d accessible port(s)", len(ports))

			for i, port := range ports {
				log.Printf("\n📡 Running tests for port %d/%d:", i+1, len(ports))
				log.Printf("   Host: %s", port.HostName)
				log.Printf("   Port: %s", port.PortNumber)
				log.Printf("   Protocol: %s", port.Protocol)
				log.Printf("   Provider Service: %s", port.ProviderServiceType)
				log.Printf("   Catalog Type: %s", port.CatalogType)

				totalTests++
				result := runPortTest(t, port, featuresPath, *outputDir)

				switch result {
				case "passed":
					passedTests++
					log.Printf("   ✅ PASSED")
				case "failed":
					failedTests++
					log.Printf("   ❌ FAILED")
				case "skipped":
					skippedTests++
					log.Printf("   ⏭️  SKIPPED")
				}
			}
		}
	} else {
		log.Println("⏭️  Skipping port tests")
	}

	// Run service tests
	if !*skipServices {
		log.Println("\n🔍 Discovering services...")
		services, err := inspection.GetServices(ctx, *provider)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to discover services: %v", err)
		} else {
			log.Printf("   Found %d service(s)", len(services))

			for i, service := range services {
				// Skip services without a catalog type
				if service.CatalogType == "" {
					log.Printf("\n⏭️  Skipping service %d/%d (no catalog type mapping):", i+1, len(services))
					log.Printf("   Provider Service: %s", service.ProviderServiceType)
					continue
				}

				log.Printf("\n🛠️  Running tests for service %d/%d:", i+1, len(services))
				log.Printf("   Resource Name: %s", service.ResourceName)
				log.Printf("   Provider Service: %s", service.ProviderServiceType)
				log.Printf("   Catalog Type: %s", service.CatalogType)
				log.Printf("   Region: %s", service.Region)
				log.Printf("   UID: %s", service.UID)

				totalTests++
				result := runServiceTest(t, service, featuresPath, *outputDir)

				switch result {
				case "passed":
					passedTests++
					log.Printf("   ✅ PASSED")
				case "failed":
					failedTests++
					log.Printf("   ❌ FAILED")
				case "skipped":
					skippedTests++
					log.Printf("   ⏭️  SKIPPED")
				}
			}
		}
	} else {
		log.Println("⏭️  Skipping service tests")
	}

	// Combine all OCSF files into a single file
	log.Println("\n🔗 Combining OCSF output files...")
	if err := combineOCSFFiles(*outputDir); err != nil {
		log.Printf("⚠️  Warning: Failed to combine OCSF files: %v", err)
	} else {
		log.Printf("   ✅ Combined OCSF file created: %s", filepath.Join(*outputDir, "combined.ocsf.json"))
	}

	// Print summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("📊 Test Summary")
	log.Printf("   Total Tests: %d", totalTests)
	log.Printf("   Passed: %d", passedTests)
	log.Printf("   Failed: %d", failedTests)
	log.Printf("   Skipped: %d", skippedTests)
	log.Println(strings.Repeat("=", 60))

	// Report final results
	if failedTests > 0 {
		log.Println("❌ Some tests failed")
		t.Fail()
	} else if totalTests == 0 {
		log.Println("⚠️  No tests were run")
		t.Fail()
	} else {
		log.Println("✅ All tests passed")
	}
}

// runPortTest runs tests for a single port configuration
func runPortTest(t *testing.T, port inspection.TestParams, featuresPath, outputDir string) string {
	// Create a safe filename from the port details
	filename := fmt.Sprintf("port-%s-%s:%s",
		sanitizeFilename(port.ResourceName),
		sanitizeFilename(port.HostName),
		sanitizeFilename(port.PortNumber),
	)
	reportPath := filepath.Join(outputDir, filename)

	// Create a subtest for this port
	result := "passed"
	t.Run(filename, func(st *testing.T) {
		cloud.RunPortTests(st, port, featuresPath, reportPath)
		if st.Failed() {
			result = "failed"
		} else if st.Skipped() {
			result = "skipped"
		}
	})

	return result
}

// runServiceTest runs tests for a single service configuration
func runServiceTest(t *testing.T, service inspection.TestParams, featuresPath, outputDir string) string {
	// Create a safe filename from the service details
	filename := fmt.Sprintf("service-%s",
		sanitizeFilename(service.ResourceName),
	)
	reportPath := filepath.Join(outputDir, filename)

	// Create a subtest for this service
	result := "passed"
	t.Run(filename, func(st *testing.T) {
		cloud.RunServiceTests(st, service, featuresPath, reportPath)
		if st.Failed() {
			result = "failed"
		} else if st.Skipped() {
			result = "skipped"
		}
	})

	return result
}

// sanitizeFilename removes characters that aren't safe for filenames
func sanitizeFilename(s string) string {
	// Replace sequences of non-alphanumeric characters with a single dash
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	s = re.ReplaceAllString(s, "-")

	// Remove leading and trailing dashes
	s = strings.Trim(s, "-")

	// Truncate if too long
	if len(s) > 100 {
		s = s[:100]
		// Remove trailing dash if truncation created one
		s = strings.TrimSuffix(s, "-")
	}

	return s
}

// combineOCSFFiles combines all *ocsf.json files in the output directory into a single combined_ocsf.json file
func combineOCSFFiles(outputDir string) error {
	// Find all OCSF JSON files in the output directory
	pattern := filepath.Join(outputDir, "*ocsf.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find OCSF files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("   No OCSF files found to combine")
		return nil
	}

	log.Printf("   Found %d OCSF file(s) to combine", len(files))

	// Combine all JSON arrays into a single array
	var combined []interface{}

	for _, file := range files {
		// Read the file
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("   ⚠️  Warning: Failed to read %s: %v", filepath.Base(file), err)
			continue
		}

		// Parse the JSON array
		var items []interface{}
		if err := json.Unmarshal(data, &items); err != nil {
			log.Printf("   ⚠️  Warning: Failed to parse %s: %v", filepath.Base(file), err)
			continue
		}

		// Add items to the combined array
		combined = append(combined, items...)
		log.Printf("   Added %d item(s) from %s", len(items), filepath.Base(file))
	}

	// Write the combined array to a new file
	combinedPath := filepath.Join(outputDir, "combined.ocsf.json")
	combinedData, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal combined data: %w", err)
	}

	if err := os.WriteFile(combinedPath, combinedData, 0644); err != nil {
		return fmt.Errorf("failed to write combined file: %w", err)
	}

	log.Printf("   Total items in combined file: %d", len(combined))

	return nil
}
