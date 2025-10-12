package cloud

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
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
	// Don't reset CloudWorld - just reset Props
	// This ensures step registrations remain valid
	suite.Props = make(map[string]interface{})

	// Setup pre-configured variables for @PerPort tests
	suite.Props["portNumber"] = params.PortNumber
	suite.Props["hostName"] = params.HostName
	suite.Props["protocol"] = params.Protocol
	suite.Props["serviceType"] = params.ServiceType
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
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create HTML output file
	htmlReportPath := reportPath + ".html"
	ocsfReportPath := reportPath + ".ocsf.json"

	// Create formatter factory with test parameters
	factory := reporters.NewFormatterFactory(params)

	// Register formatters
	godog.Format("html", "HTML report", factory.GetHTMLFormatterFunc())
	godog.Format("ocsf", "OCSF report", factory.GetOCSFFormatterFunc())

	// Build tag filter based on protocol
	tagFilter := buildTagFilter(params.Protocol)
	t.Logf("Using tag filter: %s", tagFilter)

	// Create report title
	reportTitle := "Port Test Report: " + params.HostName + ":" + params.PortNumber + " (" + params.Protocol + ")"

	opts := godog.Options{
		Format:   "html:" + htmlReportPath + ",ocsf:" + ocsfReportPath,
		Paths:    []string{featuresPath},
		Tags:     tagFilter,
		TestingT: t,
	}

	status := godog.TestSuite{
		Name: reportTitle,
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			suite.InitializeScenarioWithParams(ctx, params)
		},
		Options: &opts,
	}.Run()

	t.Logf("HTML report generated: %s", htmlReportPath)

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}
