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

// All known protocols for tag filtering
var allProtocols = []string{"http", "ssh", "smtp", "ftp", "dns", "ldap", "telnet", "mysql", "postgres", "imap", "pop3"}

// TestSuite for running cloud tests
type TestSuite struct {
	*CloudWorld
}

// NewTestSuite creates a new test suite
func NewTestSuite() *TestSuite {
	world := NewCloudWorld()
	return &TestSuite{
		CloudWorld: world,
	}
}

// Setup method called before each scenario with provided parameters
func (suite *TestSuite) setupWithParams(params reporters.TestParams) {
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

// InitializeScenarioWithParams initializes the scenario context with custom parameters
func (suite *TestSuite) InitializeScenarioWithParams(ctx *godog.ScenarioContext, params reporters.TestParams) {
	// Setup before each scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.setupWithParams(params)
		return ctx, nil
	})

	// Register all cloud steps (which includes generic steps)
	suite.RegisterSteps(ctx)
}

// buildTagFilter builds the tag expression for filtering tests based on protocol
func buildTagFilter(protocol string) string {
	// Start with @PerPort requirement
	tags := []string{"@PerPort"}

	// Build exclusion list for all other protocols (but not the current one)
	var exclusions []string
	for _, p := range allProtocols {
		if p != protocol {
			exclusions = append(exclusions, "~@"+p)
		}
	}

	// Combine: @PerPort && ~@otherProtocol1 && ~@otherProtocol2 ...
	// This means: run @PerPort tests that are NOT tagged with other protocols
	return strings.Join(append(tags, exclusions...), " && ")
}

// RunPortTests runs godog tests for a specific port configuration
func RunPortTests(t *testing.T, params reporters.TestParams, featuresPath, reportPath string) {
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
	htmlFormat := fmt.Sprintf("html-port-%p", suite)
	ocsfFormat := fmt.Sprintf("ocsf-port-%p", suite)

	godog.Format(htmlFormat, "HTML report for port tests", factory.GetHTMLFormatterFunc())
	godog.Format(ocsfFormat, "OCSF report for port tests", factory.GetOCSFFormatterFunc())

	// Build tag filter based on protocol
	tagFilter := buildTagFilter(params.Protocol)
	t.Logf("Using tag filter: %s", tagFilter)

	// Create report title
	reportTitle := "Port Test Report: " + params.HostName + ":" + params.PortNumber + " (" + params.Protocol + ")"

	opts := godog.Options{
		Format:   fmt.Sprintf("%s:%s,%s:%s", htmlFormat, htmlReportPath, ocsfFormat, ocsfReportPath),
		Paths:    []string{featuresPath},
		Tags:     tagFilter,
		TestingT: nil, // Don't use TestingT to allow proper file output
	}

	status := godog.TestSuite{
		Name: reportTitle,
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			suite.InitializeScenarioWithParams(ctx, params)
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
