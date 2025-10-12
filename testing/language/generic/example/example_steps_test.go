package main

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
)

// ExampleSteps contains example-specific step definitions
type ExampleSteps struct {
	*generic.PropsWorld
}

// APIClient is an example API client for testing
type APIClient struct {
	baseURL string
}

func (c *APIClient) Get(endpoint string) map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "success",
		"data": []interface{}{
			map[string]interface{}{
				"id":     1,
				"name":   "John Doe",
				"active": true,
			},
			map[string]interface{}{
				"id":     2,
				"name":   "Jane Doe",
				"active": false,
			},
		},
	}
}

func (c *APIClient) Post(endpoint string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":  201,
		"message": "created",
		"id":      123,
	}
}

// NewExampleSteps creates a new example steps instance
func NewExampleSteps(world *generic.PropsWorld) *ExampleSteps {
	return &ExampleSteps{
		PropsWorld: world,
	}
}

// RegisterExampleSteps registers example-specific step definitions
func (es *ExampleSteps) RegisterExampleSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I have an API client configured in "([^"]*)"$`, es.iHaveAnAPIClientConfiguredIn)
	ctx.Step(`^I have test data in "([^"]*)"$`, es.iHaveTestDataIn)
	ctx.Step(`^"([^"]*)" is a function which throws an error$`, es.functionThrowsError)
	ctx.Step(`^"([^"]*)" is a string array with colors$`, es.stringArrayWithColors)
	ctx.Step(`^"([^"]*)" is an empty array$`, es.emptyArray)
	ctx.Step(`^"([^"]*)" is an empty string$`, es.emptyString)
	ctx.Step(`^"([^"]*)" is a struct with Name "([^"]*)" and Age "([^"]*)"$`, es.structWithNameAndAge)

	// Test helper functions - one for each parameter count
	ctx.Step(`^"([^"]*)" is a test function with no parameters$`, es.functionNoParams)
	ctx.Step(`^"([^"]*)" is a test function with one parameter$`, es.functionOneParam)
	ctx.Step(`^"([^"]*)" is a test function with two parameters$`, es.functionTwoParams)
	ctx.Step(`^"([^"]*)" is a test function with three parameters$`, es.functionThreeParams)
	ctx.Step(`^I have a test object in "([^"]*)"$`, es.iHaveTestObjectIn)
}

// Example-specific step definitions
func (es *ExampleSteps) iHaveAnAPIClientConfiguredIn(variableName string) error {
	// Create and configure the API client in the specified variable
	apiClient := &APIClient{baseURL: "https://api.example.com"}
	es.Props[variableName] = apiClient
	return nil
}

func (es *ExampleSteps) iHaveTestDataIn(variableName string) error {
	// Set up test data in the specified variable
	testData := []interface{}{
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
	es.Props[variableName] = testData
	return nil
}

func (es *ExampleSteps) functionThrowsError(functionName string) error {
	// Create a function that returns an error when called
	errorFunc := func() interface{} {
		return fmt.Errorf("something went wrong")
	}
	es.Props[functionName] = errorFunc
	return nil
}

func (es *ExampleSteps) stringArrayWithColors(variableName string) error {
	// Create a string array with color values
	es.Props[variableName] = []interface{}{"red", "blue", "green"}
	return nil
}

func (es *ExampleSteps) emptyArray(variableName string) error {
	// Create an empty array
	es.Props[variableName] = []interface{}{}
	return nil
}

func (es *ExampleSteps) emptyString(variableName string) error {
	// Create an empty string
	es.Props[variableName] = ""
	return nil
}

// Test helper functions - one for each parameter count
func (es *ExampleSteps) functionNoParams(functionName string) error {
	fn := func() interface{} {
		return "no-params-result"
	}
	es.Props[functionName] = fn
	return nil
}

func (es *ExampleSteps) functionOneParam(functionName string) error {
	fn := func(a string) interface{} {
		return "one-param:" + a
	}
	es.Props[functionName] = fn
	return nil
}

func (es *ExampleSteps) functionTwoParams(functionName string) error {
	fn := func(a, b string) interface{} {
		return "two-params:" + a + "," + b
	}
	es.Props[functionName] = fn
	return nil
}

func (es *ExampleSteps) functionThreeParams(functionName string) error {
	fn := func(a, b, c string) interface{} {
		return "three-params:" + a + "," + b + "," + c
	}
	es.Props[functionName] = fn
	return nil
}

// TestObject is a simple object with methods for testing
type TestObject struct{}

func (to *TestObject) GetValue() interface{} {
	return "test-value"
}

func (to *TestObject) CombineStrings(a, b interface{}) interface{} {
	return fmt.Sprintf("%v-%v", a, b)
}

func (to *TestObject) JoinThree(a, b, c interface{}) interface{} {
	return fmt.Sprintf("%v-%v-%v", a, b, c)
}

func (es *ExampleSteps) iHaveTestObjectIn(variableName string) error {
	es.Props[variableName] = &TestObject{}
	return nil
}

// Person is a struct for testing struct field navigation
type Person struct {
	Name string
	Age  string
}

func (es *ExampleSteps) structWithNameAndAge(variableName, name, age string) error {
	person := &Person{
		Name: name,
		Age:  age,
	}
	es.Props[variableName] = person
	return nil
}
