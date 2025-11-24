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

// SetAttachmentProvider sets the attachment provider for the factory
// This allows formatters to access attachments from PropsWorld
func (ff *FormatterFactory) SetAttachmentProvider(provider attachments.Provider) {
	ff.attachmentProvider = provider
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
