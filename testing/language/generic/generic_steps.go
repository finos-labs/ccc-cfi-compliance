package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cucumber/godog"
)

// AsyncTask represents an asynchronous operation
type AsyncTask struct {
	Name      string
	StartTime time.Time
	Done      chan struct{}
	Result    interface{}
	Error     error
	Context   context.Context
	Cancel    context.CancelFunc
}

// AsyncTaskManager manages multiple async operations
type AsyncTaskManager struct {
	tasks map[string]*AsyncTask
	mutex sync.RWMutex
}

// NewAsyncTaskManager creates a new async task manager
func NewAsyncTaskManager() *AsyncTaskManager {
	return &AsyncTaskManager{
		tasks: make(map[string]*AsyncTask),
	}
}

// StartTask starts a new async task
func (atm *AsyncTaskManager) StartTask(name string, fn func(ctx context.Context) (interface{}, error)) {
	atm.mutex.Lock()
	defer atm.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	task := &AsyncTask{
		Name:      name,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Context:   ctx,
		Cancel:    cancel,
	}

	atm.tasks[name] = task

	go func() {
		defer close(task.Done)
		task.Result, task.Error = fn(ctx)
	}()
}

// WaitForTask waits for a task to complete with timeout
func (atm *AsyncTaskManager) WaitForTask(name string, timeout time.Duration) error {
	atm.mutex.RLock()
	task, exists := atm.tasks[name]
	atm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	select {
	case <-task.Done:
		return task.Error
	case <-time.After(timeout):
		task.Cancel() // Cancel the task on timeout
		return fmt.Errorf("task %s timed out after %v", name, timeout)
	}
}

// GetTaskResult gets the result of a completed task
func (atm *AsyncTaskManager) GetTaskResult(name string) (interface{}, error) {
	atm.mutex.RLock()
	defer atm.mutex.RUnlock()

	task, exists := atm.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task %s not found", name)
	}

	select {
	case <-task.Done:
		return task.Result, task.Error
	default:
		return nil, fmt.Errorf("task %s is still running", name)
	}
}

// Attachment represents a file or data attached to a test
type Attachment struct {
	Name      string
	MediaType string
	Data      []byte
}

// PropsWorld represents the test context equivalent to TypeScript PropsWorld
type PropsWorld struct {
	Props        map[string]interface{}
	T            TestingT // Interface for assertions
	AsyncManager *AsyncTaskManager
	Attachments  []Attachment // Store attachments for the current scenario
	mutex        sync.RWMutex
}

// TestingT interface for testing assertions
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// NewPropsWorld creates a new test world instance
func NewPropsWorld() *PropsWorld {
	return &PropsWorld{
		Props:        make(map[string]interface{}),
		AsyncManager: NewAsyncTaskManager(),
		Attachments:  make([]Attachment, 0),
	}
}

// Attach adds an attachment to the current scenario
func (pw *PropsWorld) Attach(name, mediaType string, data []byte) {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	pw.Attachments = append(pw.Attachments, Attachment{
		Name:      name,
		MediaType: mediaType,
		Data:      data,
	})
	fmt.Printf("📎 Attached: %s (%s, %d bytes)\n", name, mediaType, len(data))
}

// formatValueForComparison formats a value for display in comparisons
// Complex types (maps, slices, structs) are JSON-formatted, simple types are stringified
func formatValueForComparison(value interface{}) string {
	if value == nil {
		return "null"
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()

	// For complex types, use JSON formatting
	switch kind {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		if jsonBytes, err := json.MarshalIndent(value, "", "  "); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%+v", value)
	default:
		return fmt.Sprintf("%v (type: %T)", value, value)
	}
}

// HandleResolve resolves variables and literals from string references
func (pw *PropsWorld) HandleResolve(name string) interface{} {
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		stripped := name[1 : len(name)-1]

		switch stripped {
		case "nil":
			return nil
		case "true":
			return true
		case "false":
			return false
		default:
			// Try to parse as number
			if val, err := strconv.ParseFloat(stripped, 64); err == nil {
				return val
			}

			// Try direct property lookup first
			if val, exists := pw.Props[stripped]; exists {
				return val
			}

			// Try struct field access (e.g., "connection.state")
			if strings.Contains(stripped, ".") {
				parts := strings.Split(stripped, ".")
				if len(parts) == 2 {
					objName := parts[0]
					fieldName := parts[1]

					if obj, exists := pw.Props[objName]; exists {
						// Try to access the field using reflection
						v := reflect.ValueOf(obj)

						// Handle pointer types
						if v.Kind() == reflect.Ptr {
							v = v.Elem()
						}

						if v.Kind() == reflect.Struct {
							// Try the field name as-is (case-sensitive)
							field := v.FieldByName(fieldName)
							if field.IsValid() {
								return field.Interface()
							}

							// Try with capitalized first letter
							capitalizedFieldName := strings.ToUpper(fieldName[:1]) + fieldName[1:]
							field = v.FieldByName(capitalizedFieldName)
							if field.IsValid() {
								return field.Interface()
							}

							// Check if the object has a getter method for this field
							// e.g., GetState() for State field
							getterName := "Get" + capitalizedFieldName
							method := reflect.ValueOf(obj).MethodByName(getterName)
							if method.IsValid() {
								results := method.Call(nil)
								if len(results) > 0 {
									return results[0].Interface()
								}
							}
						}
					}
				}
			}

			// Try JSONPath query
			if result, err := jsonpath.Get("$."+stripped, pw.Props); err == nil {
				return result
			}

			// Return original if nothing matches
			return nil
		}
	}

	return name
}

// callFunction safely calls a function with error handling
func (pw *PropsWorld) callFunction(fn interface{}, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			pw.Props["result"] = fmt.Errorf("panic: %v", r)
		}
	}()

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		pw.Props["result"] = fmt.Errorf("not a function")
		return
	}

	// Convert args to reflect.Value
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValues[i] = reflect.ValueOf(arg)
	}

	// Call function
	results := fnValue.Call(argValues)

	// Handle results
	if len(results) > 0 {
		result := results[0].Interface()
		pw.Props["result"] = result
	}
}

// doesRowMatch checks if a data row matches the expected values
func (pw *PropsWorld) doesRowMatch(expected map[string]string, actual interface{}) (bool, string) {
	actualBytes, _ := json.Marshal(actual)
	var actualMap map[string]interface{}
	json.Unmarshal(actualBytes, &actualMap)

	var debugInfo []string
	debugInfo = append(debugInfo, fmt.Sprintf("Actual object: %s", string(actualBytes)))
	debugInfo = append(debugInfo, "Field comparisons:")

	for field, expectedVal := range expected {
		if strings.HasSuffix(field, "matches_type") {
			// Schema validation mode - would need schema validation library
			// For now, skip this validation
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: SKIPPED (schema validation)", field))
			continue
		}

		// Use JSONPath to get the field value
		var foundVal interface{}
		if result, err := jsonpath.Get("$."+field, actualMap); err == nil {
			foundVal = result
		} else {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: NOT FOUND (JSONPath error: %v)", field, err))
		}

		resolvedExpected := pw.HandleResolve(expectedVal)

		// Handle boolean comparisons
		if foundVal == true && resolvedExpected == "true" {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MATCH (true == 'true')", field))
			continue
		}
		if foundVal == false && resolvedExpected == "false" {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MATCH (false == 'false')", field))
			continue
		}

		// Convert both to strings for comparison
		foundStr := fmt.Sprintf("%v", foundVal)
		expectedStr := fmt.Sprintf("%v", resolvedExpected)

		if foundStr != expectedStr {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MISMATCH - found: '%s' (type: %T), expected: '%s' (type: %T)",
				field, foundStr, foundVal, expectedStr, resolvedExpected))
			return false, strings.Join(debugInfo, "\n")
		} else {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MATCH - '%s'", field, foundStr))
		}
	}

	return true, strings.Join(debugInfo, "\n")
}

// matchData validates array data against expected table
func (pw *PropsWorld) matchData(actual []interface{}, expected []map[string]string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("length mismatch: expected %d, got %d", len(expected), len(actual))
	}

	for i, expectedRow := range expected {
		matches, debugInfo := pw.doesRowMatch(expectedRow, actual[i])
		if !matches {
			return fmt.Errorf("row %d does not match expected values:\n%s", i, debugInfo)
		}
	}

	return nil
}

// Step Definitions

func (pw *PropsWorld) iCallFunction(fnName string) error {
	fn := pw.HandleResolve(fnName)
	pw.callFunction(fn)
	return nil
}

func (pw *PropsWorld) iCallObjectWithMethod(field, fnName string) error {
	obj := pw.HandleResolve(field)

	// Use reflection to call method
	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	// Add panic recovery for better error messages
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Error calling %s.%s(): %v", field, fnName, r)
			fmt.Printf("\n❌ %s\n", errMsg)
			pw.Props["result"] = fmt.Errorf("%s", errMsg)
		}
	}()

	results := method.Call([]reflect.Value{})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) ICallObjectWithMethodWithParameter(field, fnName, param string) error {
	obj := pw.HandleResolve(field)
	paramVal := pw.HandleResolve(param)

	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	// Add panic recovery for better error messages
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Error calling %s.%s(%v): %v", field, fnName, param, r)
			fmt.Printf("\n❌ %s\n", errMsg)
			pw.Props["result"] = fmt.Errorf("%s", errMsg)
		}
	}()

	results := method.Call([]reflect.Value{reflect.ValueOf(paramVal)})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) iCallObjectWithMethodWithTwoParameters(field, fnName, param1, param2 string) error {
	obj := pw.HandleResolve(field)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)

	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	// Add panic recovery for better error messages
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Error calling %s.%s(%v, %v): %v", field, fnName, param1, param2, r)
			fmt.Printf("\n❌ %s\n", errMsg)
			pw.Props["result"] = fmt.Errorf("%s", errMsg)
		}
	}()

	results := method.Call([]reflect.Value{
		reflect.ValueOf(param1Val),
		reflect.ValueOf(param2Val),
	})

	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) iCallObjectWithMethodWithThreeParameters(field, fnName, param1, param2, param3 string) error {
	obj := pw.HandleResolve(field)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)
	param3Val := pw.HandleResolve(param3)

	// Debug output
	fmt.Printf("\n🔍 Calling %s.%s() with:\n", field, fnName)
	fmt.Printf("   param1 (%s) = %v (type: %T)\n", param1, param1Val, param1Val)
	fmt.Printf("   param2 (%s) = %v (type: %T)\n", param2, param2Val, param2Val)
	fmt.Printf("   param3 (%s) = %v (type: %T)\n", param3, param3Val, param3Val)

	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	// Add panic recovery for better error messages
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Error calling %s.%s(%v, %v, %v): %v", field, fnName, param1, param2, param3, r)
			fmt.Printf("\n❌ %s\n", errMsg)
			pw.Props["result"] = fmt.Errorf("%s", errMsg)
		}
	}()

	results := method.Call([]reflect.Value{
		reflect.ValueOf(param1Val),
		reflect.ValueOf(param2Val),
		reflect.ValueOf(param3Val),
	})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) iCallFunctionWithParameter(fnName, param string) error {
	fn := pw.HandleResolve(fnName)
	paramVal := pw.HandleResolve(param)

	pw.callFunction(fn, paramVal)
	return nil
}

func (pw *PropsWorld) iCallFunctionWithTwoParameters(fnName, param1, param2 string) error {
	fn := pw.HandleResolve(fnName)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)

	pw.callFunction(fn, param1Val, param2Val)
	return nil
}

func (pw *PropsWorld) iCallFunctionWithThreeParameters(fnName, param1, param2, param3 string) error {
	fn := pw.HandleResolve(fnName)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)
	param3Val := pw.HandleResolve(param3)

	pw.callFunction(fn, param1Val, param2Val, param3Val)
	return nil
}

func (pw *PropsWorld) IReferToAs(from, to string) error {
	resolved := pw.HandleResolve(from)
	pw.Props[to] = resolved
	return nil
}

func (pw *PropsWorld) fieldIsSliceOfObjectsWithContents(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	// Convert to slice of interfaces
	actualSlice, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("field %s is not a slice", field)
	}

	// Convert table to expected format
	expected := make([]map[string]string, len(table.Rows)-1) // Skip header
	headers := make([]string, len(table.Rows[0].Cells))

	for i, cell := range table.Rows[0].Cells {
		headers[i] = cell.Value
	}

	for i := 1; i < len(table.Rows); i++ {
		row := make(map[string]string)
		for j, cell := range table.Rows[i].Cells {
			row[headers[j]] = cell.Value
		}
		expected[i-1] = row
	}

	return pw.matchData(actualSlice, expected)
}

func (pw *PropsWorld) fieldIsSliceOfObjectsWithAtLeastContents(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	// Convert to slice of interfaces
	actualSlice, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("field %s is not a slice", field)
	}

	// Convert table to expected format
	expected := make([]map[string]string, len(table.Rows)-1) // Skip header
	headers := make([]string, len(table.Rows[0].Cells))

	for i, cell := range table.Rows[0].Cells {
		headers[i] = cell.Value
	}

	for i := 1; i < len(table.Rows); i++ {
		row := make(map[string]string)
		for j, cell := range table.Rows[i].Cells {
			row[headers[j]] = cell.Value
		}
		expected[i-1] = row
	}

	// Check that at least the expected rows exist in the actual slice
	for _, expectedRow := range expected {
		found := false
		for _, actualItem := range actualSlice {
			match, _ := pw.doesRowMatch(expectedRow, actualItem)
			if match {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("✗ Expected row not found in slice: %+v\n", expectedRow)
			return fmt.Errorf("expected row not found: %+v", expectedRow)
		}
	}

	fmt.Printf("✓ All expected rows found in slice\n")
	return nil
}

func (pw *PropsWorld) fieldIsSliceOfObjectsWhichDoesntContainAnyOf(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	// Convert to slice of interfaces
	actualSlice, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("field %s is not a slice", field)
	}

	// Convert table to expected format
	unwanted := make([]map[string]string, len(table.Rows)-1) // Skip header
	headers := make([]string, len(table.Rows[0].Cells))

	for i, cell := range table.Rows[0].Cells {
		headers[i] = cell.Value
	}

	for i := 1; i < len(table.Rows); i++ {
		row := make(map[string]string)
		for j, cell := range table.Rows[i].Cells {
			row[headers[j]] = cell.Value
		}
		unwanted[i-1] = row
	}

	// Check that none of the unwanted rows exist in the actual slice
	for _, unwantedRow := range unwanted {
		for _, actualItem := range actualSlice {
			match, _ := pw.doesRowMatch(unwantedRow, actualItem)
			if match {
				fmt.Printf("✗ Unwanted row found in slice: %+v\n", unwantedRow)
				return fmt.Errorf("unwanted row found in slice: %+v", unwantedRow)
			}
		}
	}

	fmt.Printf("✓ None of the unwanted rows found in slice\n")
	return nil
}

func (pw *PropsWorld) fieldIsSliceOfObjectsWithLength(field, lengthField string) error {
	actual := pw.HandleResolve(field)
	expectedLength := pw.HandleResolve(lengthField)

	actualSlice, ok := actual.([]interface{})
	if !ok {
		fmt.Printf("EXPECTED: slice\n")
		fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))
		return fmt.Errorf("field %s is not a slice", field)
	}

	expectedLen, err := strconv.Atoi(fmt.Sprintf("%v", expectedLength))
	if err != nil {
		return fmt.Errorf("invalid length: %v", expectedLength)
	}

	fmt.Printf("EXPECTED: slice with length %d\n", expectedLen)
	fmt.Printf("ACTUAL:   slice with length %d\n", len(actualSlice))

	if len(actualSlice) != expectedLen {
		return fmt.Errorf("expected length %d, got %d", expectedLen, len(actualSlice))
	}

	fmt.Printf("✓ Slice length matches\n")
	return nil
}

func (pw *PropsWorld) fieldIsSliceOfStringsWithValues(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	actualSlice, ok := actual.([]interface{})
	if !ok {
		fmt.Printf("EXPECTED: slice of strings\n")
		fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))
		return fmt.Errorf("field %s is not a slice", field)
	}

	// Build expected values
	expectedValues := make([]string, len(table.Rows)-1)
	for i := 1; i < len(table.Rows); i++ {
		expectedValues[i-1] = table.Rows[i].Cells[0].Value
	}

	fmt.Printf("EXPECTED: %s\n", formatValueForComparison(expectedValues))
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	// Compare element by element
	if len(actualSlice) != len(expectedValues) {
		return fmt.Errorf("slice length mismatch: expected %d, got %d", len(expectedValues), len(actualSlice))
	}

	for i, expectedVal := range expectedValues {
		actualVal := fmt.Sprintf("%v", actualSlice[i])
		if actualVal != expectedVal {
			fmt.Printf("EXPECTED[%d]: %s\n", i, expectedVal)
			fmt.Printf("ACTUAL[%d]:   %s\n", i, actualVal)
			return fmt.Errorf("element %d mismatch: expected %s, got %s", i, expectedVal, actualVal)
		}
	}

	fmt.Printf("✓ All elements match\n")
	return nil
}

func (pw *PropsWorld) fieldIsObjectWithContents(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	// Convert table to expected format (should have only one data row)
	if len(table.Rows) != 2 {
		return fmt.Errorf("expected exactly one data row in table")
	}

	expected := make(map[string]string)
	for i, cell := range table.Rows[0].Cells {
		expected[cell.Value] = table.Rows[1].Cells[i].Value
	}

	fmt.Printf("EXPECTED: %s\n", formatValueForComparison(expected))
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	// Check each expected field
	actualMap, ok := actual.(map[string]interface{})
	if !ok {
		return fmt.Errorf("field %s is not an object/map", field)
	}

	for key, expectedVal := range expected {
		actualVal, exists := actualMap[key]
		if !exists {
			fmt.Printf("EXPECTED[%s]: %s\n", key, expectedVal)
			fmt.Printf("ACTUAL[%s]:   <missing>\n", key)
			return fmt.Errorf("field %s missing in actual object", key)
		}

		actualStr := fmt.Sprintf("%v", actualVal)
		if actualStr != expectedVal {
			fmt.Printf("EXPECTED[%s]: %s\n", key, expectedVal)
			fmt.Printf("ACTUAL[%s]:   %s\n", key, actualStr)
			return fmt.Errorf("field %s mismatch: expected %s, got %s", key, expectedVal, actualStr)
		}
	}

	fmt.Printf("✓ All fields match\n")
	return nil
}

func (pw *PropsWorld) fieldIsNil(field string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: null\n")
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	if actual != nil {
		return fmt.Errorf("expected %s to be nil, got %v", field, actual)
	}

	fmt.Printf("✓ Value is nil\n")
	return nil
}

func (pw *PropsWorld) fieldIsNotNil(field string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: not null\n")
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	if actual == nil {
		return fmt.Errorf("expected %s to not be nil", field)
	}

	fmt.Printf("✓ Value is not nil\n")
	return nil
}

func (pw *PropsWorld) fieldIsTrue(field string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: true (type: bool)\n")
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	if actual != true {
		return fmt.Errorf("expected %s to be true, got %v", field, actual)
	}

	fmt.Printf("✓ Value is true\n")
	return nil
}

func (pw *PropsWorld) fieldIsFalse(field string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: false (type: bool)\n")
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	if actual != false {
		return fmt.Errorf("expected %s to be false, got %v", field, actual)
	}

	fmt.Printf("✓ Value is false\n")
	return nil
}

func (pw *PropsWorld) fieldIsEmpty(field string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: empty\n")
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	switch v := actual.(type) {
	case []interface{}:
		if len(v) != 0 {
			return fmt.Errorf("expected %s to be empty, got length %d", field, len(v))
		}
		fmt.Printf("✓ Array is empty\n")
	case string:
		if len(v) != 0 {
			return fmt.Errorf("expected %s to be empty, got length %d", field, len(v))
		}
		fmt.Printf("✓ String is empty\n")
	default:
		return fmt.Errorf("cannot check if %s is empty: unsupported type", field)
	}

	return nil
}

func (pw *PropsWorld) fieldEquals(field, expected string) error {
	actual := pw.HandleResolve(field)
	expectedVal := pw.HandleResolve(expected)

	fmt.Printf("EXPECTED: %s\n", formatValueForComparison(expectedVal))
	fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))

	if !reflect.DeepEqual(actual, expectedVal) {
		// Try string conversion for comparison
		actualStr := fmt.Sprintf("%v", actual)
		expectedStr := fmt.Sprintf("%v", expectedVal)

		if actualStr == expectedStr {
			fmt.Printf("✓ Values match (after string conversion)\n")
			return nil
		}

		return fmt.Errorf("expected %s to equal %s, got %s", field, expectedStr, actualStr)
	}

	fmt.Printf("✓ Values match\n")
	return nil
}

func (pw *PropsWorld) fieldIsErrorWithMessage(field, errorType string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: error with message \"%s\"\n", errorType)

	if err, ok := actual.(error); ok {
		fmt.Printf("ACTUAL:   error with message \"%s\"\n", err.Error())

		if err.Error() != errorType {
			return fmt.Errorf("expected error message %s, got %s", errorType, err.Error())
		}

		fmt.Printf("✓ Error message matches\n")
	} else {
		fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))
		return fmt.Errorf("expected %s to be an error", field)
	}

	return nil
}

func (pw *PropsWorld) fieldIsError(field string) error {
	actual := pw.HandleResolve(field)

	fmt.Printf("EXPECTED: error\n")

	if err, ok := actual.(error); ok {
		fmt.Printf("ACTUAL:   error: %s\n", err.Error())
		fmt.Printf("✓ Value is an error\n")
	} else {
		fmt.Printf("ACTUAL:   %s\n", formatValueForComparison(actual))
		return fmt.Errorf("expected %s to be an error, got %T", field, actual)
	}

	return nil
}

func (pw *PropsWorld) HandlerIsInvocationCounter(handlerName, field string) error {
	pw.Props[handlerName] = func() {
		count, exists := pw.Props[field]
		if !exists {
			count = 0
		}
		if countInt, ok := count.(int); ok {
			pw.Props[field] = countInt + 1
		} else {
			pw.Props[field] = 1
		}
	}
	pw.Props[field] = 0
	return nil
}

func (pw *PropsWorld) FunctionReturnsPromiseOf(fnName, field string) error {
	value := pw.HandleResolve(field)
	pw.Props[fnName] = func() interface{} {
		return value
	}
	return nil
}

func (pw *PropsWorld) waitForPeriod(ms string) error {
	duration, err := strconv.Atoi(ms)
	if err != nil {
		return fmt.Errorf("invalid duration: %s", ms)
	}

	time.Sleep(time.Duration(duration) * time.Millisecond)
	return nil
}

// Async-specific step definitions

// iStartTaskByCallingFunction starts a task by calling a function with 0 parameters
func (pw *PropsWorld) iStartTaskByCallingFunction(taskName, functionName string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the function directly without writing to shared result
		funcValue := pw.HandleResolve(functionName)
		if funcValue == nil {
			return nil, fmt.Errorf("function %s not found", functionName)
		}

		fn, ok := funcValue.(func() interface{})
		if !ok {
			return nil, fmt.Errorf("%s is not a callable function", functionName)
		}

		result := fn()
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingFunctionWithParameter starts a task by calling a function with 1 parameter
func (pw *PropsWorld) iStartTaskByCallingFunctionWithParameter(taskName, functionName, param1 string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the function directly without writing to shared result
		funcValue := pw.HandleResolve(functionName)
		if funcValue == nil {
			return nil, fmt.Errorf("function %s not found", functionName)
		}

		resolvedParam1 := pw.HandleResolve(param1)
		fn, ok := funcValue.(func(string) interface{})
		if !ok {
			return nil, fmt.Errorf("%s is not a callable function with 1 parameter", functionName)
		}

		result := fn(fmt.Sprintf("%v", resolvedParam1))
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingFunctionWithTwoParameters starts a task by calling a function with 2 parameters
func (pw *PropsWorld) iStartTaskByCallingFunctionWithTwoParameters(taskName, functionName, param1, param2 string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the function directly without writing to shared result
		funcValue := pw.HandleResolve(functionName)
		if funcValue == nil {
			return nil, fmt.Errorf("function %s not found", functionName)
		}

		resolvedParam1 := pw.HandleResolve(param1)
		resolvedParam2 := pw.HandleResolve(param2)
		fn, ok := funcValue.(func(string, string) interface{})
		if !ok {
			return nil, fmt.Errorf("%s is not a callable function with 2 parameters", functionName)
		}

		result := fn(fmt.Sprintf("%v", resolvedParam1), fmt.Sprintf("%v", resolvedParam2))
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingFunctionWithThreeParameters starts a task by calling a function with 3 parameters
func (pw *PropsWorld) iStartTaskByCallingFunctionWithThreeParameters(taskName, functionName, param1, param2, param3 string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the function directly without writing to shared result
		funcValue := pw.HandleResolve(functionName)
		if funcValue == nil {
			return nil, fmt.Errorf("function %s not found", functionName)
		}

		resolvedParam1 := pw.HandleResolve(param1)
		resolvedParam2 := pw.HandleResolve(param2)
		resolvedParam3 := pw.HandleResolve(param3)
		fn, ok := funcValue.(func(string, string, string) interface{})
		if !ok {
			return nil, fmt.Errorf("%s is not a callable function with 3 parameters", functionName)
		}

		result := fn(fmt.Sprintf("%v", resolvedParam1), fmt.Sprintf("%v", resolvedParam2), fmt.Sprintf("%v", resolvedParam3))
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingObjectWithMethod starts a task by calling an object method with 0 parameters
func (pw *PropsWorld) iStartTaskByCallingObjectWithMethod(taskName, objectName, methodName string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the method directly without writing to shared result
		obj := pw.HandleResolve(objectName)
		if obj == nil {
			return nil, fmt.Errorf("object %s not found", objectName)
		}

		val := reflect.ValueOf(obj)
		method := val.MethodByName(methodName)
		if !method.IsValid() {
			return nil, fmt.Errorf("method %s not found on object %s", methodName, objectName)
		}

		results := method.Call([]reflect.Value{})
		if len(results) == 0 {
			return nil, nil
		}

		result := results[0].Interface()
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingObjectWithMethodWithParameter starts a task by calling an object method with 1 parameter
func (pw *PropsWorld) iStartTaskByCallingObjectWithMethodWithParameter(taskName, objectName, methodName, param1 string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the method directly without writing to shared result
		obj := pw.HandleResolve(objectName)
		if obj == nil {
			return nil, fmt.Errorf("object %s not found", objectName)
		}

		val := reflect.ValueOf(obj)
		method := val.MethodByName(methodName)
		if !method.IsValid() {
			return nil, fmt.Errorf("method %s not found on object %s", methodName, objectName)
		}

		resolvedParam1 := pw.HandleResolve(param1)
		results := method.Call([]reflect.Value{reflect.ValueOf(resolvedParam1)})
		if len(results) == 0 {
			return nil, nil
		}

		result := results[0].Interface()
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingObjectWithMethodWithTwoParameters starts a task by calling an object method with 2 parameters
func (pw *PropsWorld) iStartTaskByCallingObjectWithMethodWithTwoParameters(taskName, objectName, methodName, param1, param2 string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the method directly without writing to shared result
		obj := pw.HandleResolve(objectName)
		if obj == nil {
			return nil, fmt.Errorf("object %s not found", objectName)
		}

		val := reflect.ValueOf(obj)
		method := val.MethodByName(methodName)
		if !method.IsValid() {
			return nil, fmt.Errorf("method %s not found on object %s", methodName, objectName)
		}

		resolvedParam1 := pw.HandleResolve(param1)
		resolvedParam2 := pw.HandleResolve(param2)
		results := method.Call([]reflect.Value{reflect.ValueOf(resolvedParam1), reflect.ValueOf(resolvedParam2)})
		if len(results) == 0 {
			return nil, nil
		}

		result := results[0].Interface()
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// iStartTaskByCallingObjectWithMethodWithThreeParameters starts a task by calling an object method with 3 parameters
func (pw *PropsWorld) iStartTaskByCallingObjectWithMethodWithThreeParameters(taskName, objectName, methodName, param1, param2, param3 string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Call the method directly without writing to shared result
		obj := pw.HandleResolve(objectName)
		if obj == nil {
			return nil, fmt.Errorf("object %s not found", objectName)
		}

		val := reflect.ValueOf(obj)
		method := val.MethodByName(methodName)
		if !method.IsValid() {
			return nil, fmt.Errorf("method %s not found on object %s", methodName, objectName)
		}

		resolvedParam1 := pw.HandleResolve(param1)
		resolvedParam2 := pw.HandleResolve(param2)
		resolvedParam3 := pw.HandleResolve(param3)
		results := method.Call([]reflect.Value{reflect.ValueOf(resolvedParam1), reflect.ValueOf(resolvedParam2), reflect.ValueOf(resolvedParam3)})
		if len(results) == 0 {
			return nil, nil
		}

		result := results[0].Interface()
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	})
	return nil
}

// All-in-one wait functions (start task and wait for completion)

// iWaitForFunction calls a function and waits for it to complete
func (pw *PropsWorld) iWaitForFunction(functionName string) error {
	taskName := "temp_" + functionName
	err := pw.iStartTaskByCallingFunction(taskName, functionName)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForFunctionWithParameter calls a function with 1 parameter and waits for it
func (pw *PropsWorld) iWaitForFunctionWithParameter(functionName, param1 string) error {
	taskName := "temp_" + functionName
	err := pw.iStartTaskByCallingFunctionWithParameter(taskName, functionName, param1)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForFunctionWithTwoParameters calls a function with 2 parameters and waits for it
func (pw *PropsWorld) iWaitForFunctionWithTwoParameters(functionName, param1, param2 string) error {
	taskName := "temp_" + functionName
	err := pw.iStartTaskByCallingFunctionWithTwoParameters(taskName, functionName, param1, param2)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForFunctionWithThreeParameters calls a function with 3 parameters and waits for it
func (pw *PropsWorld) iWaitForFunctionWithThreeParameters(functionName, param1, param2, param3 string) error {
	taskName := "temp_" + functionName
	err := pw.iStartTaskByCallingFunctionWithThreeParameters(taskName, functionName, param1, param2, param3)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForObjectWithMethod calls an object method and waits for it
func (pw *PropsWorld) iWaitForObjectWithMethod(objectName, methodName string) error {
	taskName := "temp_" + objectName + "_" + methodName
	err := pw.iStartTaskByCallingObjectWithMethod(taskName, objectName, methodName)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForObjectWithMethodWithParameter calls an object method with 1 parameter and waits
func (pw *PropsWorld) iWaitForObjectWithMethodWithParameter(objectName, methodName, param1 string) error {
	taskName := "temp_" + objectName + "_" + methodName
	err := pw.iStartTaskByCallingObjectWithMethodWithParameter(taskName, objectName, methodName, param1)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForObjectWithMethodWithTwoParameters calls an object method with 2 parameters and waits
func (pw *PropsWorld) iWaitForObjectWithMethodWithTwoParameters(objectName, methodName, param1, param2 string) error {
	taskName := "temp_" + objectName + "_" + methodName
	err := pw.iStartTaskByCallingObjectWithMethodWithTwoParameters(taskName, objectName, methodName, param1, param2)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForObjectWithMethodWithThreeParameters calls an object method with 3 parameters and waits
func (pw *PropsWorld) iWaitForObjectWithMethodWithThreeParameters(objectName, methodName, param1, param2, param3 string) error {
	taskName := "temp_" + objectName + "_" + methodName
	err := pw.iStartTaskByCallingObjectWithMethodWithThreeParameters(taskName, objectName, methodName, param1, param2, param3)
	if err != nil {
		return err
	}
	return pw.iWaitForTaskToComplete(taskName)
}

// iWaitForTaskToComplete waits for a task to complete (default 30s timeout)
func (pw *PropsWorld) iWaitForTaskToComplete(taskName string) error {
	err := pw.AsyncManager.WaitForTask(taskName, 30*time.Second)
	if err != nil {
		pw.Props["result"] = err
		return nil
	}

	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err != nil {
		pw.Props["result"] = err
	} else {
		pw.Props["result"] = result
	}
	return nil
}

// iWaitForTaskToCompleteWithinMs waits for a task to complete within a specified millisecond timeout
func (pw *PropsWorld) iWaitForTaskToCompleteWithinMs(taskName, timeoutMs string) error {
	timeoutVal, err := strconv.Atoi(timeoutMs)
	if err != nil {
		return fmt.Errorf("invalid timeout: %s", timeoutMs)
	}

	timeout := time.Duration(timeoutVal) * time.Millisecond
	err = pw.AsyncManager.WaitForTask(taskName, timeout)
	if err != nil {
		pw.Props["result"] = err
		return nil
	}

	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err != nil {
		pw.Props["result"] = err
	} else {
		pw.Props["result"] = result
	}
	return nil
}

// RegisterSteps registers all step definitions with the Godog suite
func (pw *PropsWorld) RegisterSteps(s *godog.ScenarioContext) {

	// Function call patterns - direct functions
	s.Step(`^I call "([^"]*)"$`, pw.iCallFunction)

	// Function call patterns - object methods
	s.Step(`^I call "([^"]*)" with "([^"]*)"$`, pw.iCallObjectWithMethod)
	s.Step(`^I call "([^"]*)" with "([^"]*)" with parameter "([^"]*)"$`, pw.ICallObjectWithMethodWithParameter)
	s.Step(`^I call "([^"]*)" with "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iCallObjectWithMethodWithTwoParameters)
	s.Step(`^I call "([^"]*)" with "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iCallObjectWithMethodWithThreeParameters)

	// Function call patterns - direct functions
	s.Step(`^I call "([^"]*)" with parameter "([^"]*)"$`, pw.iCallFunctionWithParameter)
	s.Step(`^I call "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iCallFunctionWithTwoParameters)
	s.Step(`^I call "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iCallFunctionWithThreeParameters)

	// Variable management
	s.Step(`^I refer to "([^"]*)" as "([^"]*)"$`, pw.IReferToAs)

	// Data validation patterns
	s.Step(`^"([^"]*)" is a slice of objects with the following contents$`, pw.fieldIsSliceOfObjectsWithContents)
	s.Step(`^"([^"]*)" is a slice of objects with at least the following contents$`, pw.fieldIsSliceOfObjectsWithAtLeastContents)
	s.Step(`^"([^"]*)" is a slice of objects which doesn't contain any of$`, pw.fieldIsSliceOfObjectsWhichDoesntContainAnyOf)
	s.Step(`^"([^"]*)" is a slice of objects with length "([^"]*)"$`, pw.fieldIsSliceOfObjectsWithLength)
	s.Step(`^"([^"]*)" is a slice of strings with the following values$`, pw.fieldIsSliceOfStringsWithValues)
	s.Step(`^"([^"]*)" is an object with the following contents$`, pw.fieldIsObjectWithContents)

	// Value assertions
	s.Step(`^"([^"]*)" is nil$`, pw.fieldIsNil)
	s.Step(`^"([^"]*)" is not nil$`, pw.fieldIsNotNil)
	s.Step(`^"([^"]*)" is true$`, pw.fieldIsTrue)
	s.Step(`^"([^"]*)" is false$`, pw.fieldIsFalse)
	s.Step(`^"([^"]*)" is empty$`, pw.fieldIsEmpty)
	s.Step(`^"([^"]*)" is "([^"]*)"$`, pw.fieldEquals)
	s.Step(`^"([^"]*)" is an error with message "([^"]*)"$`, pw.fieldIsErrorWithMessage)
	s.Step(`^"([^"]*)" is an error$`, pw.fieldIsError)

	// Test setup patterns
	s.Step(`^"([^"]*)" is a invocation counter into "([^"]*)"$`, pw.HandlerIsInvocationCounter)
	s.Step(`^"([^"]*)" is a function which returns a value of "([^"]*)"$`, pw.FunctionReturnsPromiseOf)
	s.Step(`^we wait for a period of "([^"]*)" ms$`, pw.waitForPeriod)

	// Async task patterns - starting tasks
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)"$`, pw.iStartTaskByCallingFunction)
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with parameter "([^"]*)"$`, pw.iStartTaskByCallingFunctionWithParameter)
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iStartTaskByCallingFunctionWithTwoParameters)
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iStartTaskByCallingFunctionWithThreeParameters)

	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with "([^"]*)"$`, pw.iStartTaskByCallingObjectWithMethod)
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with "([^"]*)" with parameter "([^"]*)"$`, pw.iStartTaskByCallingObjectWithMethodWithParameter)
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iStartTaskByCallingObjectWithMethodWithTwoParameters)
	s.Step(`^I start task "([^"]*)" by calling "([^"]*)" with "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iStartTaskByCallingObjectWithMethodWithThreeParameters)

	// Async task patterns - waiting for tasks
	s.Step(`^I wait for task "([^"]*)" to complete$`, pw.iWaitForTaskToComplete)
	s.Step(`^I wait for task "([^"]*)" to complete within "([^"]*)" ms$`, pw.iWaitForTaskToCompleteWithinMs)

	// Async task patterns - all-in-one (start and wait)
	s.Step(`^I wait for "([^"]*)"$`, pw.iWaitForFunction)
	s.Step(`^I wait for "([^"]*)" with parameter "([^"]*)"$`, pw.iWaitForFunctionWithParameter)
	s.Step(`^I wait for "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iWaitForFunctionWithTwoParameters)
	s.Step(`^I wait for "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iWaitForFunctionWithThreeParameters)

	s.Step(`^I wait for "([^"]*)" with "([^"]*)"$`, pw.iWaitForObjectWithMethod)
	s.Step(`^I wait for "([^"]*)" with "([^"]*)" with parameter "([^"]*)"$`, pw.iWaitForObjectWithMethodWithParameter)
	s.Step(`^I wait for "([^"]*)" with "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iWaitForObjectWithMethodWithTwoParameters)
	s.Step(`^I wait for "([^"]*)" with "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iWaitForObjectWithMethodWithThreeParameters)
}
