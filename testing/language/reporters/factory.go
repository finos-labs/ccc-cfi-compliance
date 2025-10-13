package reporters

import (
	"io"

	"github.com/cucumber/godog/formatters"
	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
)

// TestParams is an alias to inspection.TestParams for backward compatibility
type TestParams = inspection.TestParams

// FormatterFactory creates formatters with embedded test parameters
type FormatterFactory struct {
	params TestParams
}

// NewFormatterFactory creates a new formatter factory with the given parameters
func NewFormatterFactory(params TestParams) *FormatterFactory {
	return &FormatterFactory{
		params: params,
	}
}

// UpdateParams updates the test parameters for this factory
// Call this before running each test to ensure formatters use the correct params
func (ff *FormatterFactory) UpdateParams(params TestParams) {
	ff.params = params
}

// GetHTMLFormatterFunc returns a configured HTML formatter function
func (ff *FormatterFactory) GetHTMLFormatterFunc() func(string, io.Writer) formatters.Formatter {
	return func(suite string, out io.Writer) formatters.Formatter {
		return NewHTMLFormatterWithParams(suite, out, ff.params)
	}
}

// GetOCSFFormatterFunc returns a configured OCSF formatter function
func (ff *FormatterFactory) GetOCSFFormatterFunc() func(string, io.Writer) formatters.Formatter {
	return func(suite string, out io.Writer) formatters.Formatter {
		return NewOCSFFormatterWithParams(suite, out, ff.params)
	}
}
