package main

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// TestSuite for running godog features
type TestSuite struct {
	*generic.PropsWorld
	ExampleSteps *ExampleSteps
}

// NewTestSuite creates a new test suite
func NewTestSuite() *TestSuite {
	world := generic.NewPropsWorld()
	return &TestSuite{
		PropsWorld:   world,
		ExampleSteps: NewExampleSteps(world),
	}
}

// Setup method called before each scenario
func (suite *TestSuite) setup() {
	// Reset the world for each scenario
	suite.PropsWorld = generic.NewPropsWorld()
	suite.ExampleSteps = NewExampleSteps(suite.PropsWorld)

	// Setup test data for examples
	suite.Props["apiClient"] = &APIClient{baseURL: "https://api.example.com"}
	suite.Props["testData"] = map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}
	suite.Props["users"] = []interface{}{
		map[string]interface{}{
			"name":   "John Doe",
			"active": true,
			"profile": map[string]interface{}{
				"email": "john@example.com",
			},
		},
		map[string]interface{}{
			"name":   "Jane Doe",
			"active": false,
			"profile": map[string]interface{}{
				"email": "jane@example.com",
			},
		},
	}
}

// InitializeScenario initializes the scenario context for godog
func (suite *TestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	// Setup before each scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.setup()
		return ctx, nil
	})

	// Register all generic steps
	suite.RegisterSteps(ctx)

	// Register example-specific steps
	suite.ExampleSteps.RegisterExampleSteps(ctx)
}

// Global function for godog CLI (required by godog)
func InitializeScenario(ctx *godog.ScenarioContext) {
	suite := NewTestSuite()
	suite.InitializeScenario(ctx)
}

// Test function for running with go test
func TestGodogFeatures(t *testing.T) {
	suite := NewTestSuite()
	suite.T = t

	factory := reporters.NewFormatterFactory(reporters.TestParams{})

	godog.Format("html", "HTML report", factory.GetHTMLFormatterFunc())

	opts := godog.Options{
		Format:   "html:cucumber-report.html",
		Paths:    []string{"features"},
		TestingT: t,
	}

	status := godog.TestSuite{
		Name:                "Generic Steps Examples",
		ScenarioInitializer: suite.InitializeScenario,
		Options:             &opts,
	}.Run()

	t.Log("HTML report generated: cucumber-report.html")

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}
