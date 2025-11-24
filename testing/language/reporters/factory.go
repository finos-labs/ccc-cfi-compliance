package reporters

import (
	"io"

	"github.com/cucumber/godog/formatters"
	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/attachments"
)

// TestParams is an alias to inspection.TestParams for backward compatibility
type TestParams = inspection.TestParams

// FormatterFactory creates formatters with embedded test parameters
type FormatterFactory struct {
	params             TestParams
	attachmentProvider attachments.Provider
}

// NewFormatterFactory creates a new formatter factory with the given parameters
// Optionally accepts an attachment provider as the second parameter
func NewFormatterFactory(params TestParams, attachmentProvider ...attachments.Provider) *FormatterFactory {
	ff := &FormatterFactory{
		params: params,
	}
	if len(attachmentProvider) > 0 {
		ff.attachmentProvider = attachmentProvider[0]
	}
	return ff
}

// UpdateParams updates the test parameters for this factory
// Call this before running each test to ensure formatters use the correct params
func (ff *FormatterFactory) UpdateParams(params TestParams) {
	ff.params = params
}

// GetHTMLFormatterFunc returns a configured HTML formatter function
func (ff *FormatterFactory) GetHTMLFormatterFunc() func(string, io.Writer) formatters.Formatter {
	return func(suite string, out io.Writer) formatters.Formatter {
		return NewHTMLFormatterWithAttachments(suite, out, ff.params, ff.attachmentProvider)
	}
}

// GetOCSFFormatterFunc returns a configured OCSF formatter function
func (ff *FormatterFactory) GetOCSFFormatterFunc() func(string, io.Writer) formatters.Formatter {
	return func(suite string, out io.Writer) formatters.Formatter {
		return NewOCSFFormatterWithParams(suite, out, ff.params)
	}
}
