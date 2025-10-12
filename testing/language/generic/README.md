# Generic Cucumber Steps for Go

This package provides a generic set of Cucumber/Godog step definitions for testing in Go.

## Features

- Variable assignment and resolution
- Function calls with parameters
- Object and array validation
- Boolean and nil checks
- Error handling
- Asynchronous operations with timeouts
- Data table support for complex assertions
- Detailed EXPECTED/ACTUAL logging for all comparisons

## Running Tests

### Run all tests

```bash
cd example
go test -v
```

HTML report is generated at: `cucumber-report.html`

### Run tests with coverage

```bash
# Run from the generic directory to measure coverage of generic_steps.go
go test -v -coverprofile=coverage.out -coverpkg=. ./example
```

### View coverage report in terminal

```bash
# See coverage for generic_steps.go functions
go tool cover -func=coverage.out | grep generic_steps

# Or see all coverage
go tool cover -func=coverage.out
```

### View coverage report in browser

```bash
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Usage

See `example/example.feature` for comprehensive examples of all available step definitions.

## Documentation

See [Cucumber-Generic-Language.md](./Cucumber-Generic-Language.md) for complete documentation of all available step definitions and patterns.
