package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/factory"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// TestSuite for running cloud tests
type TestSuite struct {
	*cloud.CloudWorld
}

// NewTestSuite creates a new test suite
func NewTestSuite() *TestSuite {
	world := cloud.NewCloudWorld()
	return &TestSuite{
		CloudWorld: world,
	}
}

// setupServiceParams sets up parameters for service tests
// Accepts any struct and populates Props using reflection
func (suite *TestSuite) setupServiceParams(params any) {
	// Use reflection to automatically populate all fields from the params struct
	v := reflect.ValueOf(params)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Only process if it's a struct
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		suite.Props[field.Name] = value.Interface() // Use .Interface() to get the actual value, not reflect.Value
	}
}

// InitializeServiceScenario initializes the scenario context for service testing
func (suite *TestSuite) InitializeServiceScenario(sc *godog.ScenarioContext, params environment.TestParams) {
	// Setup before each scenario
	sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		suite.Props = make(map[string]interface{})
		suite.AsyncManager = generic.NewAsyncTaskManager()
		suite.setupServiceParams(params)
		suite.setupServiceParams(params.CloudParams)
		return ctx, nil
	})

	// Register all cloud steps (which includes generic steps)
	suite.RegisterSteps(sc)
}

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

	// Discover resources using GetOrProvisionTestableResources
	log.Println("🔍 Discovering testable resources...")
	resources, err := service.GetOrProvisionTestableResources()
	if err != nil {
		log.Fatalf("Failed to discover resources: %v", err)
	}

	if len(resources) > 0 {
		if resourcesJSON, err := json.MarshalIndent(resources, "   ", "  "); err == nil {
			log.Printf("   Resources:\n   %s", string(resourcesJSON))
		}
	}
	log.Println()

	// Get features directory path and collect all subdirectories
	_, filename, _, _ := runtime.Caller(0)
	runnerDir := filepath.Dir(filename)
	testingDir := filepath.Dir(runnerDir)
	featuresBaseDir := filepath.Join(testingDir, "features")

	// Collect all catalog subdirectories (CCC.ObjStor, CCC.Core, etc.)
	featuresPaths := []string{}
	entries, err := os.ReadDir(featuresBaseDir)
	if err != nil {
		log.Fatalf("Failed to read features directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			featuresPaths = append(featuresPaths, filepath.Join(featuresBaseDir, entry.Name()))
		}
	}

	log.Printf("📂 Features Paths: %s", strings.Join(featuresPaths, ", "))
	log.Println()

	// Run tests for each resource
	stats := r.runTests(ctx, resources, featuresPaths)

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

// runTests executes tests for all resources
func (r *BasicServiceRunner) runTests(ctx context.Context, resources []environment.TestParams, featuresPaths []string) TestStats {
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
		result := r.runResourceTest(ctx, resource, featuresPaths, resource.CatalogTypes)

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
func (r *BasicServiceRunner) runResourceTest(ctx context.Context, params environment.TestParams, featuresPaths []string, catalogTypes []string) string {
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
	suite := NewTestSuite()

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
	log.Printf("   Tag Filter: %s", tagFilter)

	opts := godog.Options{
		Format:      fmt.Sprintf("%s:%s,%s:%s", htmlFormat, htmlReportPath, ocsfFormat, ocsfReportPath),
		Paths:       featuresPaths,
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
	return strings.Join(catalogTypes, ",")
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
