package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/factory"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// BasicServiceRunner provides functionality for running service-specific compliance tests
type BasicServiceRunner struct {
	Config RunConfig
}

// NewBasicServiceRunner creates a new basic service runner
func NewBasicServiceRunner(config RunConfig) *BasicServiceRunner {
	return &BasicServiceRunner{
		Config: config,
	}
}

// GetConfig returns the run configuration
func (r *BasicServiceRunner) GetConfig() RunConfig {
	return r.Config
}

// Run executes the compliance tests (implements ServiceRunner interface)
func (r *BasicServiceRunner) Run() int {
	config := r.Config

	log.Printf("🚀 Starting CCC Compliance Tests")
	log.Printf("   Service: %s", config.ServiceName)
	log.Printf("   Provider: %s", config.CloudParams.Provider)
	log.Println()

	// Discover features path based on service name
	featuresPath, err := r.discoverFeaturesPath(config.ServiceName)
	if err != nil {
		log.Fatalf("Failed to discover features path for service '%s': %v", config.ServiceName, err)
	}
	log.Printf("📁 Features Path: %s", featuresPath)
	log.Println()

	// Clean and create output directory
	if err := r.prepareOutputDirectory(config.OutputDir); err != nil {
		log.Fatalf("Failed to prepare output directory: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Create cloud factory
	cloudFactory, err := factory.NewFactory(factory.CloudProvider(config.CloudParams.Provider), config.CloudParams)
	if err != nil {
		log.Fatalf("Failed to create factory: %v", err)
	}

	// Get the service from the factory
	log.Printf("🔧 Getting service: %s", config.ServiceName)
	service, err := cloudFactory.GetServiceAPI(config.ServiceName)
	if err != nil {
		log.Fatalf("Failed to get service '%s': %v", config.ServiceName, err)
	}

	// Discover resources using GetTestableResources
	log.Println("🔍 Discovering testable resources...")
	resources, err := service.GetTestableResources()
	if err != nil {
		log.Fatalf("Failed to discover resources: %v", err)
	}

	log.Printf("   Found %d resource(s)", len(resources))

	// Run tests for each resource
	stats := r.runTests(ctx, resources, featuresPath)

	// Print summary
	r.printSummary(stats)

	// Return exit code
	if stats.Failed > 0 {
		return 1
	} else if stats.Total == 0 {
		return 1
	}
	return 0
}

// TestStats tracks test execution statistics
type TestStats struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

// discoverFeaturesPath discovers the features directory for a service
func (r *BasicServiceRunner) discoverFeaturesPath(serviceName string) (string, error) {
	// Map service names to CCC catalog directories
	serviceMap := map[string]string{
		"object-storage": "CCC.ObjStor",
		"iam":            "CCC.IAM",
	}

	catalogDir, ok := serviceMap[serviceName]
	if !ok {
		return "", fmt.Errorf("unknown service name: %s", serviceName)
	}

	// Get the path relative to this file
	_, filename, _, _ := runtime.Caller(0)
	runnerDir := filepath.Dir(filename)                                           // testing/runner/
	testingDir := filepath.Dir(runnerDir)                                         // testing/
	featuresPath := filepath.Join(testingDir, "services", catalogDir, "features") // testing/services/CCC.ObjStor/features/

	// Check if directory exists
	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		return "", fmt.Errorf("features directory does not exist: %s", featuresPath)
	}

	return featuresPath, nil
}

// prepareOutputDirectory cleans and creates the output directory
func (r *BasicServiceRunner) prepareOutputDirectory(outputDir string) error {
	log.Printf("🧹 Cleaning output directory: %s", outputDir)
	if err := os.RemoveAll(outputDir); err != nil && !os.IsNotExist(err) {
		log.Printf("⚠️  Warning: Failed to clean output directory: %v", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	log.Printf("✅ Output directory ready")
	log.Println()
	return nil
}

// displayResourcesCSV prints resources in CSV format
func (r *BasicServiceRunner) displayResourcesCSV(resources []environment.TestParams) {
	log.Println("\n# Resources:")
	log.Println("Provider,ResourceName,UID,ProviderServiceType,CatalogTypes,Region")
	for _, res := range resources {
		log.Printf("%s,%s,%s,%s,%s,%s",
			res.CloudParams.Provider,
			res.ResourceName,
			res.UID,
			res.ProviderServiceType,
			strings.Join(res.CatalogTypes, "|"),
			res.CloudParams.Region)
	}
	log.Println()
}

// runTests executes tests for all resources
func (r *BasicServiceRunner) runTests(ctx context.Context, resources []environment.TestParams, featuresPath string) TestStats {
	stats := TestStats{}

	for i, resource := range resources {
		// Skip resources that don't match the filter
		if r.Config.ResourceFilter != "" && resource.ResourceName != r.Config.ResourceFilter {
			continue
		}

		log.Printf("\n🔬 Running tests for resource %d/%d:", i+1, len(resources))
		if resourceJSON, err := json.MarshalIndent(resource, "   ", "  "); err == nil {
			log.Printf("   Resource: %s", string(resourceJSON))
		} else {
			log.Printf("   Resource: %+v", resource)
		}

		stats.Total++
		result := r.runResourceTest(ctx, resource, featuresPath, resource.CatalogTypes)

		switch result {
		case "passed":
			stats.Passed++
			log.Printf("   ✅ PASSED")
		case "failed":
			stats.Failed++
			log.Printf("   ❌ FAILED")
		case "skipped":
			stats.Skipped++
			log.Printf("   ⏭️  SKIPPED")
		}
	}

	return stats
}

// runResourceTest runs tests for a single resource
func (r *BasicServiceRunner) runResourceTest(ctx context.Context, params environment.TestParams, featuresPath string, catalogTypes []string) string {
	// Create a safe filename from the resource name
	filename := fmt.Sprintf("resource-%s", sanitizeFilename(params.ResourceName))
	reportPath := filepath.Join(r.Config.OutputDir, filename)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(reportPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Failed to create output directory: %v", err)
		return "failed"
	}

	// Run the godog tests
	suite := cloud.NewTestSuite()

	// Create HTML and OCSF output files
	htmlReportPath := reportPath + ".html"
	ocsfReportPath := reportPath + ".ocsf.json"

	// Create formatter factory
	formatterFactory := reporters.NewFormatterFactory(params, suite.CloudWorld)

	// Generate unique format names
	htmlFormat := fmt.Sprintf("html-resource-%s", sanitizeFilename(params.ResourceName))
	ocsfFormat := fmt.Sprintf("ocsf-resource-%s", sanitizeFilename(params.ResourceName))

	godog.Format(htmlFormat, "HTML report", formatterFactory.GetHTMLFormatterFunc())
	godog.Format(ocsfFormat, "OCSF report", formatterFactory.GetOCSFFormatterFunc())

	// Build tag filter
	tagFilter := buildTagFilter(params.CatalogTypes)

	opts := godog.Options{
		Format:      fmt.Sprintf("%s:%s,%s:%s", htmlFormat, htmlReportPath, ocsfFormat, ocsfReportPath),
		Paths:       []string{featuresPath},
		Tags:        tagFilter,
		Concurrency: 1,
		Strict:      true,
		NoColors:    false,
	}

	status := godog.TestSuite{
		Name: fmt.Sprintf("%s Test: %s", strings.Join(catalogTypes, "/"), params.ResourceName),
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			suite.InitializeServiceScenario(sc, params)
		},
		Options: &opts,
	}.Run()

	// Determine result
	if status == 0 {
		return "passed"
	} else if status == 2 {
		return "skipped"
	}
	return "failed"
}

// buildTagFilter builds tag expression for filtering tests
func buildTagFilter(catalogTypes []string) string {
	// Build OR expression for catalog types: @CCC.ObjStor || @CCC.Core
	var catalogTags []string
	for _, ct := range catalogTypes {
		catalogTags = append(catalogTags, "@"+ct)
	}
	catalogExpr := strings.Join(catalogTags, " || ")

	// Require @PerService AND one of the catalog types
	return fmt.Sprintf("@PerService && (%s)", catalogExpr)
}

// printSummary prints test execution summary
func (r *BasicServiceRunner) printSummary(stats TestStats) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("📊 Test Summary")
	log.Printf("   Total Tests: %d", stats.Total)
	log.Printf("   Passed: %d", stats.Passed)
	log.Printf("   Failed: %d", stats.Failed)
	log.Printf("   Skipped: %d", stats.Skipped)
	log.Println(strings.Repeat("=", 60))

	if stats.Failed > 0 {
		log.Println("❌ Some tests failed")
	} else if stats.Total == 0 {
		log.Println("⚠️  No tests were run")
	} else {
		log.Println("✅ All tests passed")
	}
}

// sanitizeFilename removes characters that aren't safe for filenames
func sanitizeFilename(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		} else {
			result += "-"
		}
	}
	return result
}
