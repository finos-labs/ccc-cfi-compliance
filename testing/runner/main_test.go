package runner

import (
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

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/factory"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
)

const (
	// featuresPath is hardcoded relative to this file (testing/runner/main_test.go)
	// Features are in testing/features/
	featuresPath = "../features"
)

var (
	provider    = flag.String("provider", "", "Cloud provider (aws, azure, or gcp)")
	outputDir   = flag.String("output", "output", "Output directory for test reports")
	timeout     = flag.Duration("timeout", 30*time.Minute, "Timeout for all tests")
	serviceType = flag.String("service-type", "", "Optional: Run tests only for a specific service type (e.g., object-storage, iam)")
)

func TestRunCompliance(t *testing.T) {

	// Validate required flags
	if *provider == "" {
		log.Fatal("Error: -provider flag is required (aws, azure, or gcp)")
	}

	if *provider != "aws" && *provider != "azure" && *provider != "gcp" {
		log.Fatalf("Error: invalid provider '%s' (must be aws, azure, or gcp)", *provider)
	}

	log.Printf("🚀 Starting CCC CFI Compliance Tests")
	log.Printf("   Provider: %s", *provider)
	log.Printf("   Output Directory: %s", *outputDir)
	log.Printf("   Features Path: %s", featuresPath)
	log.Printf("   Timeout: %s", *timeout)
	if *serviceType != "" {
		log.Printf("   Service Type Filter: %s", *serviceType)
	}
	log.Println()

	// Clean and create output directory
	log.Printf("🧹 Cleaning output directory: %s", *outputDir)
	if err := os.RemoveAll(*outputDir); err != nil && !os.IsNotExist(err) {
		log.Printf("⚠️  Warning: Failed to clean output directory: %v", err)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	log.Printf("✅ Output directory ready")
	log.Println()

	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedTests := 0

	// Create factory for the cloud provider
	cloudProvider := factory.CloudProvider(*provider)
	f, err := factory.NewFactory(cloudProvider)
	if err != nil {
		log.Fatalf("❌ Failed to create factory: %v", err)
	}

	// Determine which service types to test
	serviceTypes := []factory.ServiceType{}
	if *serviceType != "" {
		// Test only the specified service type
		serviceTypes = append(serviceTypes, factory.ServiceType(*serviceType))
	} else {
		// Test all implemented service types
		serviceTypes = append(serviceTypes, factory.ServiceTypeObjectStorage)
		// Add more as they are implemented: factory.ServiceTypeIAM, etc.
	}

	// Iterate through each service type
	for _, svcType := range serviceTypes {
		log.Printf("\n🔧 Testing service type: %s", svcType)

		// Get the service API for this type using the typed method
		svc, err := f.GetServiceAPIForType(svcType)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to get service API for %s: %v", svcType, err)
			continue
		}

		// Check if the service implements the generic.Service interface
		service, ok := svc.(generic.Service)
		if !ok {
			log.Printf("   ⚠️  Service type %s does not implement generic.Service interface", svcType)
			continue
		}

		// Run tests using the generic service
		stats := runServiceTestsGeneric(t, service, svcType, featuresPath, *outputDir)
		totalTests += stats.total
		passedTests += stats.passed
		failedTests += stats.failed
		skippedTests += stats.skipped
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

// testStats holds statistics from running tests
type testStats struct {
	total   int
	passed  int
	failed  int
	skipped int
}

// runServiceTestsGeneric runs tests for all instances of a service
func runServiceTestsGeneric(t *testing.T, service generic.Service, serviceType factory.ServiceType, featuresPath, outputDir string) testStats {
	stats := testStats{}

	// Get all instances as TestParams
	log.Println("   🔍 Discovering instances...")
	instances, err := service.GetAllInstances()
	if err != nil {
		log.Printf("   ⚠️  Warning: Failed to get instances: %v", err)
		return stats
	}
	log.Printf("   Found %d instance(s)", len(instances))

	// Run tests for each instance
	for i, params := range instances {
		log.Printf("\n   📦 Running tests for instance %d/%d:", i+1, len(instances))
		log.Printf("      Resource: %s", params.ResourceName)
		log.Printf("      UID: %s", params.UID)
		log.Printf("      Region: %s", params.Region)
		log.Printf("      Catalog: %s", params.CatalogType)

		stats.total++
		result := runInstanceTest(t, params, serviceType, featuresPath, outputDir)

		switch result {
		case "passed":
			stats.passed++
			log.Printf("      ✅ PASSED")
		case "failed":
			stats.failed++
			log.Printf("      ❌ FAILED")
		case "skipped":
			stats.skipped++
			log.Printf("      ⏭️  SKIPPED")
		}
	}

	return stats
}

// runInstanceTest runs tests for a single service instance
func runInstanceTest(t *testing.T, params inspection.TestParams, serviceType factory.ServiceType, featuresPath, outputDir string) string {
	// Create a safe filename from the resource name
	filename := fmt.Sprintf("%s-%s", serviceType, sanitizeFilename(params.ResourceName))
	reportPath := filepath.Join(outputDir, filename)

	// Create a subtest for this instance
	result := "passed"
	t.Run(filename, func(st *testing.T) {
		// Run the actual godog tests for this service instance
		cloud.RunServiceTests(st, params, featuresPath, reportPath)

		// Check test result
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
