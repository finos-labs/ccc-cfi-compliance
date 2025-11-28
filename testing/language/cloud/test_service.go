package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// buildServiceTagFilter builds the tag expression for filtering tests based on catalog type
func buildServiceTagFilter(catalogTypes []string) string {
	// Build OR expression for catalog types: @CCC.ObjStor || @CCC.Core
	var catalogTags []string
	for _, ct := range catalogTypes {
		catalogTags = append(catalogTags, "@"+ct)
	}
	catalogExpr := strings.Join(catalogTags, " || ")

	// Require @PerService AND one of the catalog types
	return fmt.Sprintf("@PerService && (%s)", catalogExpr)
}

// setupServiceParams sets up parameters for @PerService tests
func (suite *TestSuite) setupServiceParams(params reporters.TestParams) {
	// Reset Props and AsyncManager, but keep the same CloudWorld instance
	// so the formatter can access attachments from the previous scenario
	suite.Props = make(map[string]interface{})
	suite.AsyncManager = generic.NewAsyncTaskManager()

	// Use reflection to automatically populate all fields from TestParams
	v := reflect.ValueOf(params)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		suite.Props[field.Name] = value.Interface()
	}
}

// InitializeServiceScenario initializes the scenario context for service testing
func (suite *TestSuite) InitializeServiceScenario(ctx *godog.ScenarioContext, params reporters.TestParams) {
	// Setup before each scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.setupServiceParams(params)
		return ctx, nil
	})

	// Register all cloud steps (which includes generic steps)
	suite.RegisterSteps(ctx)
}

// RunServiceTests runs godog tests for a specific service configuration
func RunServiceTests(t *testing.T, params reporters.TestParams, featuresPath, reportPath string) {
	suite := NewTestSuite()

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(reportPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create HTML output file
	htmlReportPath := reportPath + ".html"
	ocsfReportPath := reportPath + ".ocsf.json"

	// Create formatter factory with params and attachment provider
	factory := reporters.NewFormatterFactory(params, suite.CloudWorld)

	// Generate unique format names for this test run to avoid conflicts
	htmlFormat := fmt.Sprintf("html-service-%p", suite)
	ocsfFormat := fmt.Sprintf("ocsf-service-%p", suite)

	godog.Format(htmlFormat, "HTML report for service tests", factory.GetHTMLFormatterFunc())
	godog.Format(ocsfFormat, "OCSF report for service tests", factory.GetOCSFFormatterFunc())

	// Build tag filter based on catalog types
	tagFilter := buildServiceTagFilter(params.CatalogTypes)
	t.Logf("Using tag filter: %s", tagFilter)

	// Create report title
	catalogTypesStr := strings.Join(params.CatalogTypes, "/")
	reportTitle := "Service Test Report: " + params.ResourceName + " (" + catalogTypesStr + " / " + params.ProviderServiceType + ")"

	opts := godog.Options{
		Format:   fmt.Sprintf("%s:%s,%s:%s", htmlFormat, htmlReportPath, ocsfFormat, ocsfReportPath),
		Paths:    []string{featuresPath},
		Tags:     tagFilter,
		TestingT: nil, // Don't use TestingT to allow proper file output
	}

	status := godog.TestSuite{
		Name: reportTitle,
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			suite.InitializeServiceScenario(ctx, params)
		},
		Options: &opts,
	}.Run()

	// Map godog status to testing behavior
	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.FailNow()
	}
}
