# Go Cucumber Generic Language Patterns

This document describes the main patterns and step definitions available in the Go-based cucumber testing framework used for CCC (Common Cloud Controls) compliance testing.

## Overview

The `generic_steps.go` file provides a comprehensive set of reusable step definitions that create a mini-language for testing various scenarios. This framework is designed to leverage Go's concurrency primitives (goroutines and channels) and supports both synchronous and asynchronous operations, data validation, and complex object interactions.

## Core Concepts

### Variable Resolution System

The framework uses a sophisticated variable resolution system through the `HandleResolve` method:

- **Direct values**: Simple strings are used as-is
- **Variable references**: Variables wrapped in `{}` are resolved from the test context
- **Special literals**: `{null}`, `{true}`, `{false}`, and numeric values in `{}` are converted to their respective Go types
- **JSONPath expressions**: Complex object paths can be accessed using JSONPath syntax within `{}`

### Test Context (PropsWorld)

All test data is stored in a `PropsWorld` struct that maintains state across steps:

- `pw.Props[key]` - stores and retrieves test data using Go's `map[string]interface{}`
- Results from operations are typically stored in `pw.Props["result"]`
- Thread-safe access is provided through `sync.RWMutex` for concurrent operations

## Step Definition Patterns

### 1. Reflection-Based Function Call Patterns

#### Method Calls on Go Structs

```gherkin
When I call "{struct}" with "methodName"
When I call "{struct}" with "methodName" with parameter "{variable}"
When I call "{struct}" with "methodName" with parameters "{param1}" and "{param2}"
When I call "{struct}" with "methodName" with parameters "{param1}" and "string value" and "{param3}"
```

- Uses Go's `reflect` package to call methods on structs dynamically
- Supports 0-3 parameters with automatic type conversion
- Method names must be exported (capitalized) to be accessible
- Results are stored in `result` field
- Handles method not found errors gracefully

#### Direct Function Calls

```gherkin
When I call "{function}"
When I call "{function}" with parameter "{param1}"
When I call "{function}" with parameters "{param1}" and "{param2}"
When I call "{function}" with parameters "{param1}", "{param2}" and "{param3}"
```

- Calls `func` types stored in the test context
- Supports 1-3 parameters with interface{} conversion
- Uses reflection to match parameter types dynamically
- Results are stored in `result` field

### 2. Asynchronous Operation Patterns

#### Starting A Function in a Goroutine

```gherkin
When I start task "name" by calling "{function}"
When I start task "name" by calling "{function}" with parameter "{param1}"
When I start task "name" by calling "{function}" with parameters "{param1}" and "{param2}"
When I start task "name" by calling "{function}" with parameters "{param1}", "{param2}" and "{param3}"
```

- Creates a goroutine, storing it in the props array under the value 'name'
- Supports 1-3 parameters with interface{} conversion
- Uses reflection to match parameter types dynamically

#### Starting A Method Call in a Goroutine

```gherkin
When I start task "name" by calling "{struct}" with "methodName"
When I start task "name" by calling "{struct}" with "methodName" with parameter "{param1}"
When I start task "name" by calling "{struct}" with "methodName" with parameters "{param1}" and "{param2}"
When I start task "name" by calling "{struct}" with "methodName" with parameters "{param1}", "{param2}" and "{param3}" and "{param3}"
```

- Creates a goroutine, storing it in a struct in the props array under the value 'name'
- Supports 1-3 parameters with interface{} conversion
- Uses reflection to match parameter types dynamically

#### Waiting for Completion

```gherkin
Then I wait for task "blah" to complete
Then I wait for task "blah" to complete within "time" ms
```

- A blocking wait on continuing the test.
- Uses `select` statements with timeout channels for non-blocking operations
- Cancels long-running goroutines using `context.CancelFunc`
- Puts the channel's output into `result`.

#### All In One

```gherkin
When I wait for "{function}"
When I wait for "{function}" with parameter "{param1}"
When I wait for "{function}" with parameters "{param1}" and "{param2}"
When I wait for "{function}" with parameters "{param1}", "{param2}" and "{param3}"
When I wait for "{struct}" with "methodName"
When I wait for "{struct}" with "methodName" with parameter "{variable}"
When I wait for "{struct}" with "methodName" with parameters "{param1}" and "{param2}"
When I wait for "{struct}" with "methodName" with parameters "{param1}" and "string value" and "{param3}"
```

This waits for a task to complete, so rolling two of the above steps together, effectively.

### 3. Go Variable Management Patterns

#### Variable Assignment and Aliasing

```gherkin
When I refer to "{sourceField}" as "targetField"
```

- Creates an alias or deep copies a value from one field to another
- Works with all Go types including slices, maps, and structs
- Uses Go's type system to maintain type safety where possible
- Useful for creating readable variable names and intermediate results

### 4. Go Data Structure Validation Patterns

#### Slice of Maps Validation

```gherkin
Then "{fieldName}" is a slice of objects with the following contents
| property1 | property2 | property3 |
| value1    | value2    | value3    |
| value4    | value5    | value6    |
```

- Validates that a `[]interface{}` contains `map[string]interface{}` with specific properties
- Uses Godog's data tables for structured comparison
- Supports JSONPath expressions for nested property access
- Handles Go type assertions safely

#### Slice Length Validation

```gherkin
Then "{fieldName}" is an slice of objects with length "expectedLength"
```

- Validates the length of Go slices (`[]interface{}`, `[]string`, etc.)
- Uses `len()` function for accurate slice length checking

#### String Slice Validation

```gherkin
Then "{fieldName}" is an slice of strings with the following values
| value |
| str1  |
| str2  |
| str3  |
```

- Validates `[]string` or `[]interface{}` containing strings
- Performs type assertions to ensure string types

#### Map Property Validation

```gherkin
Then "{fieldName}" is an object with the following contents
| property1 | property2 | property3 |
| value1    | value2    | value3    |
```

```gherkin
Then "{users}" is an slice of objects with the following contents
| name     | active | profile.email        |
| John Doe | true   | john@example.com     |
| Jane Doe | false  | jane@example.com     |
```

```gherkin
Then "{users}" is a slice of objects with at least the following contents
| name     | active | profile.email        |
| John Doe | true   | john@example.com     |
```

```gherkin
Then "{users}" is a slice of objects which doesn't contain any of
| name     | active | profile.email        |
| Intruder 1 | true   | int@example.com    |
| Dodgy Guy | true | dodgy@example.com     |
```

- Validates that a `map[string]interface{}` has specific key-value pairs
- Supports nested map access using JSONPath syntax
- Handles Go's zero values appropriately

### 5. Go Type Assertion Patterns

#### Nil Checks

```gherkin
Then "{fieldName}" is nil
Then "{fieldName}" is not nil
```

- `nil` checks for Go `nil` values (pointers, slices, maps, channels, functions)
- `not nil` validates that a value is not nil
- `undefined` checks if a key exists in the Props map
- Handles Go's typed nil values correctly

#### Boolean Type Assertions

```gherkin
Then "{fieldName}" is true
Then "{fieldName}" is false
```

- Performs type assertion to `bool` type
- Handles interface{} to bool conversion safely

#### Empty Collection Checks

```gherkin
Then "{fieldName}" is empty
```

- Uses Go's `len()` function for slices, maps, strings, and channels
- Validates that collections have length 0
- Supports all Go collection types

#### String Equality Comparison

```gherkin
Then "{fieldName}" is "{expectedValue}"
```

- Converts both values to strings using `fmt.Sprintf("%v", value)`
- Handles Go's string representation of various types
- Supports variable resolution in expected values

#### Go Error Interface Validation

```gherkin
Then "{fieldName}" is an error
Then "{fieldName}" is an error with message "{expectedMessage}"
```

- Validates that a field implements Go's `error` interface
- Uses type assertion `value.(error)` for error checking
- Compares error messages using `Error()` method

### 6. Go Test Setup and Concurrency Patterns

#### Goroutine-Safe Counter Functions

```gherkin
Given "{handlerName}" is a invocation counter into "{counterField}"
```

- Creates a closure that increments a counter using atomic operations
- Thread-safe for concurrent goroutine access
- Useful for tracking function call frequency in concurrent scenarios

#### Function Mock Creation

```gherkin
Given "{functionName}" is a function which returns a value of "{value}"
```

- Creates a `func() interface{}` that returns a specific value
- Supports both synchronous and context-aware function signatures
- Useful for mocking Go interfaces and function types

#### Goroutine Timing Control

```gherkin
Given we wait for a period of "{milliseconds}" ms
```

- Uses `time.Sleep()` to introduce delays in test execution
- Supports millisecond precision timing
- Useful for testing goroutine synchronization and timing-sensitive operations

### 7. Async Task Management Patterns

#### Goroutine Task Execution

```gherkin
When I start async task "{taskName}"
Then I wait for async task "{taskName}" to complete
Then I wait for async task "{taskName}" to complete within "{timeout}" seconds
```

- Manages long-running goroutines with proper lifecycle control
- Uses `context.Context` for cancellation and timeout handling
- Provides task status monitoring and result collection

#### Concurrent Task Coordination

```gherkin
Then all async tasks should complete
Then async task "{taskName}" should be running
Then async task "{taskName}" should be completed
```

- Coordinates multiple goroutines using `sync.WaitGroup` patterns
- Monitors task states using channels and select statements
- Handles task cancellation and cleanup

## 8. Go Error Handling and Panic Recovery

All step definitions include comprehensive Go error handling:

- Panics are recovered and converted to errors using `recover()`
- Errors implement Go's `error` interface and are stored in the `result` field
- Goroutine panics are handled gracefully with proper cleanup
- Provides detailed stack traces for debugging
