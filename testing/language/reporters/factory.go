package reporters

import (
	"io"

	"github.com/cucumber/godog/formatters"
)

// TestParams holds the parameters for port / service testing
type TestParams struct {
	PortNumber  string
	HostName    string
	Protocol    string
	ServiceType string
	Region      string
	Provider    string
	Labels      []string
	UID         string
}

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
