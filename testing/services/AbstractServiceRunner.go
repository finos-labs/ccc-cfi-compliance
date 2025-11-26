package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/factory"
	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// GetTestResourcesFunc is a function that discovers resources to test
type GetTestResourcesFunc func(ctx context.Context, provider string, cloudFactory factory.Factory) ([]inspection.TestParams, error)

// AbstractServiceRunner provides common functionality for running service-specific compliance tests
type AbstractServiceRunner struct {
	CatalogType      string               // e.g., "CCC.ObjStor", "CCC.Core"
	GetTestResources GetTestResourcesFunc // Function to discover resources
	Config           RunConfig
	FeaturesPath     string
}

// NewAbstractServiceRunner creates a new abstract service runner
func NewAbstractServiceRunner(catalogType string, getTestResources GetTestResourcesFunc, config RunConfig, featuresPath string) *AbstractServiceRunner {
	return &AbstractServiceRunner{
		CatalogType:      catalogType,
		GetTestResources: getTestResources,
		Config:           config,
		FeaturesPath:     featuresPath,
	}
}

// GetConfig returns the run configuration
func (r *AbstractServiceRunner) GetConfig() RunConfig {
	return r.Config
}

// Run executes the compliance tests (implements ServiceRunner interface)
func (r *AbstractServiceRunner) Run() int {
	config := r.Config
	// Validate provider
	if config.Provider == "" {
		log.Fatal("Error: provider is required (aws, azure, or gcp)")
	}

	if config.Provider != "aws" && config.Provider != "azure" && config.Provider != "gcp" {
		log.Fatalf("Error: invalid provider '%s' (must be aws, azure, or gcp)", config.Provider)
	}

	log.Printf("🚀 Starting CCC Compliance Tests")
	log.Printf("   Provider: %s", config.Provider)
	log.Printf("   Catalog Type: %s", r.CatalogType)
	log.Printf("   Output Directory: %s", config.OutputDir)
	log.Printf("   Features Path: %s", r.FeaturesPath)
	log.Printf("   Timeout: %s", config.Timeout)
	if config.ResourceFilter != "" {
		log.Printf("   Resource Filter: %s", config.ResourceFilter)
	}
	log.Println()

	// Clean and create output directory
	if err := r.prepareOutputDirectory(config.OutputDir); err != nil {
		log.Fatalf("Failed to prepare output directory: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Create cloud factory
	cloudFactory, err := factory.NewFactory(factory.CloudProvider(config.Provider))
	if err != nil {
		log.Fatalf("Failed to create factory: %v", err)
	}

	// Discover resources
	log.Println("🔍 Discovering resources...")
	resources, err := r.GetTestResources(ctx, config.Provider, cloudFactory)
	if err != nil {
		log.Fatalf("Failed to discover resources: %v", err)
	}

	log.Printf("   Found %d resource(s)", len(resources))

	// Display resources as CSV
	if len(resources) > 0 {
		r.displayResourcesCSV(resources)
	}

	// Run tests for each resource
	stats := r.runTests(ctx, resources, config)

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

// prepareOutputDirectory cleans and creates the output directory
func (r *AbstractServiceRunner) prepareOutputDirectory(outputDir string) error {
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
func (r *AbstractServiceRunner) displayResourcesCSV(resources []inspection.TestParams) {
	log.Println("\n# Resources:")
	log.Println("Provider,ResourceName,UID,ProviderServiceType,CatalogType,Region")
	for _, res := range resources {
		log.Printf("%s,%s,%s,%s,%s,%s",
			res.Provider,
			res.ResourceName,
			res.UID,
			res.ProviderServiceType,
			res.CatalogType,
			res.Region)
	}
	log.Println()
}

// runTests executes tests for all resources
func (r *AbstractServiceRunner) runTests(ctx context.Context, resources []inspection.TestParams, config RunConfig) TestStats {
	stats := TestStats{}

	for i, resource := range resources {
		// Skip resources that don't match the filter
		if config.ResourceFilter != "" && resource.ResourceName != config.ResourceFilter {
			continue
		}

		log.Printf("\n🔬 Running tests for resource %d/%d:", i+1, len(resources))
		log.Printf("   Name: %s", resource.ResourceName)
		log.Printf("   UID: %s", resource.UID)
		log.Printf("   Region: %s", resource.Region)

		stats.Total++
		result := r.runResourceTest(ctx, resource, config)

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
func (r *AbstractServiceRunner) runResourceTest(ctx context.Context, params inspection.TestParams, config RunConfig) string {
	// Create a safe filename from the resource name
	filename := fmt.Sprintf("resource-%s", sanitizeFilename(params.ResourceName))
	reportPath := filepath.Join(config.OutputDir, filename)

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
	tagFilter := buildTagFilter(params.CatalogType)

	opts := godog.Options{
		Format:      fmt.Sprintf("%s:%s,%s:%s", htmlFormat, htmlReportPath, ocsfFormat, ocsfReportPath),
		Paths:       []string{r.FeaturesPath},
		Tags:        tagFilter,
		Concurrency: 1,
		Strict:      true,
		NoColors:    false,
	}

	status := godog.TestSuite{
		Name: fmt.Sprintf("%s Test: %s", r.CatalogType, params.ResourceName),
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
func buildTagFilter(catalogType string) string {
	tags := []string{"@PerService"}

	// Add exclusions for other catalog types
	var exclusions []string
	for _, ct := range inspection.AllCatalogTypes {
		if ct != catalogType {
			exclusions = append(exclusions, "~@"+ct)
		}
	}

	return strings.Join(append(tags, exclusions...), " && ")
}

// printSummary prints test execution summary
func (r *AbstractServiceRunner) printSummary(stats TestStats) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("📊 Test Summary - %s", r.CatalogType)
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
