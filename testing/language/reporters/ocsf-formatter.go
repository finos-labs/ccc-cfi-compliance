package reporters

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
)

// OCSFFormatter is a godog formatter that generates OCSF JSON reports
type OCSFFormatter struct {
	out             io.Writer
	findings        []OCSFFinding
	currentFeature  string
	currentScenario *OCSFFinding
	scenarioStarted bool
	startTime       time.Time
	params          *PortTestParams // Optional test parameters
}

// OCSFFinding represents a single OCSF finding/result
type OCSFFinding struct {
	Message      string          `json:"message"`
	Metadata     OCSFMetadata    `json:"metadata"`
	SeverityID   int             `json:"severity_id"`
	Severity     string          `json:"severity"`
	Status       string          `json:"status"`
	StatusCode   string          `json:"status_code"`
	StatusDetail string          `json:"status_detail"`
	StatusID     int             `json:"status_id"`
	Unmapped     OCSFUnmapped    `json:"unmapped"`
	ActivityName string          `json:"activity_name"`
	ActivityID   int             `json:"activity_id"`
	FindingInfo  OCSFFindingInfo `json:"finding_info"`
	CategoryName string          `json:"category_name"`
	CategoryUID  int             `json:"category_uid"`
	ClassName    string          `json:"class_name"`
	ClassUID     int             `json:"class_uid"`
	Time         int64           `json:"time"`
	TimeDT       string          `json:"time_dt"`
	TypeUID      int             `json:"type_uid"`
	TypeName     string          `json:"type_name"`
}

// OCSFMetadata represents the metadata section
type OCSFMetadata struct {
	EventCode string      `json:"event_code"`
	Product   OCSFProduct `json:"product"`
	Profiles  []string    `json:"profiles"`
	Version   string      `json:"version"`
}

// OCSFProduct represents the product information
type OCSFProduct struct {
	Name       string `json:"name"`
	UID        string `json:"uid"`
	VendorName string `json:"vendor_name"`
	Version    string `json:"version"`
}

// OCSFUnmapped represents the unmapped section
type OCSFUnmapped struct {
	Compliance map[string][]string `json:"compliance"`
}

// OCSFFindingInfo represents the finding_info section
type OCSFFindingInfo struct {
	CreatedTime   int64    `json:"created_time"`
	CreatedTimeDT string   `json:"created_time_dt"`
	Desc          string   `json:"desc"`
	Title         string   `json:"title"`
	Types         []string `json:"types"`
	UID           string   `json:"uid"`
}

// Feature captures feature information
func (f *OCSFFormatter) Feature(gd *messages.GherkinDocument, uri string, c []byte) {
	if gd.Feature != nil {
		f.currentFeature = gd.Feature.Name
	}
}

// Pickle captures pickle (scenario) information
func (f *OCSFFormatter) Pickle(pickle *messages.Pickle) {
	// Initialize a new finding for this scenario
	now := time.Now()
	f.currentScenario = &OCSFFinding{
		Message: pickle.Name,
		Metadata: OCSFMetadata{
			EventCode: "ccc_compliance_test",
			Product: OCSFProduct{
				Name:       "CCC-Destructive",
				UID:        "CCC-Destructive",
				VendorName: "FINOS",
				Version:    "0.1",
			},
			Profiles: []string{"cloud", "datetime"},
			Version:  "1.4.0",
		},
		SeverityID: 1,
		Severity:   "Informational",
		Status:     "New",
		StatusCode: "PASS", // Default to PASS, will be updated if any step fails
		StatusID:   1,
		Unmapped: OCSFUnmapped{
			Compliance: map[string][]string{
				"CCC": {f.currentFeature},
			},
		},
		ActivityName: "Test",
		ActivityID:   1,
		FindingInfo: OCSFFindingInfo{
			CreatedTime:   now.Unix(),
			CreatedTimeDT: now.Format(time.RFC3339),
			Desc:          fmt.Sprintf("Compliance test scenario: %s", pickle.Name),
			Title:         pickle.Name,
			Types:         []string{},
			UID:           fmt.Sprintf("ccc-test-%s-%d", pickle.Id, now.Unix()),
		},
		CategoryName: "Findings",
		CategoryUID:  2,
		ClassName:    "Compliance Finding",
		ClassUID:     2004,
		Time:         now.Unix(),
		TimeDT:       now.Format(time.RFC3339),
		TypeUID:      200401,
		TypeName:     "Compliance Finding: Test",
	}
	f.scenarioStarted = true
}

// TestRunStarted is required by the formatters.Formatter interface
func (f *OCSFFormatter) TestRunStarted() {
	f.startTime = time.Now()
}

// TestRunFinished captures test run completion
func (f *OCSFFormatter) TestRunFinished(msg *messages.TestRunFinished) {
	// Finalize any pending scenario
	if f.scenarioStarted && f.currentScenario != nil {
		f.findings = append(f.findings, *f.currentScenario)
		f.scenarioStarted = false
	}
}

// Summary generates and writes the final OCSF JSON report
func (f *OCSFFormatter) Summary() {
	// Finalize any pending scenario
	if f.scenarioStarted && f.currentScenario != nil {
		f.findings = append(f.findings, *f.currentScenario)
		f.scenarioStarted = false
	}

	// Marshal to JSON with pretty printing
	jsonData, err := json.MarshalIndent(f.findings, "", "    ")
	if err != nil {
		fmt.Fprintf(f.out, "Error generating OCSF report: %v\n", err)
		return
	}

	fmt.Fprint(f.out, string(jsonData))
}

// Defined is required by the formatters.Formatter interface
func (f *OCSFFormatter) Defined(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	// No action needed
}

// Passed is required by the formatters.Formatter interface
func (f *OCSFFormatter) Passed(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	// Append step to status detail
	if f.currentScenario != nil {
		if f.currentScenario.StatusDetail != "" {
			f.currentScenario.StatusDetail += "\n"
		}
		f.currentScenario.StatusDetail += fmt.Sprintf("✓ %s", step.Text)
	}
}

// Skipped is required by the formatters.Formatter interface
func (f *OCSFFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.currentScenario != nil {
		f.currentScenario.StatusCode = "SKIP"
		if f.currentScenario.StatusDetail != "" {
			f.currentScenario.StatusDetail += "\n"
		}
		f.currentScenario.StatusDetail += fmt.Sprintf("⊘ %s (skipped)", step.Text)
	}
}

// Undefined is required by the formatters.Formatter interface
func (f *OCSFFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.currentScenario != nil {
		f.currentScenario.StatusCode = "FAIL"
		f.currentScenario.SeverityID = 3
		f.currentScenario.Severity = "Medium"
		if f.currentScenario.StatusDetail != "" {
			f.currentScenario.StatusDetail += "\n"
		}
		f.currentScenario.StatusDetail += fmt.Sprintf("? %s (undefined)", step.Text)
	}
}

// Failed is required by the formatters.Formatter interface
func (f *OCSFFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition, err error) {
	if f.currentScenario != nil {
		f.currentScenario.StatusCode = "FAIL"
		f.currentScenario.SeverityID = 3
		f.currentScenario.Severity = "Medium"
		if f.currentScenario.StatusDetail != "" {
			f.currentScenario.StatusDetail += "\n"
		}
		if err != nil {
			f.currentScenario.StatusDetail += fmt.Sprintf("✗ %s - Error: %s", step.Text, err.Error())
		} else {
			f.currentScenario.StatusDetail += fmt.Sprintf("✗ %s", step.Text)
		}
	}
}

// Pending is required by the formatters.Formatter interface
func (f *OCSFFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.currentScenario != nil {
		f.currentScenario.StatusCode = "PENDING"
		if f.currentScenario.StatusDetail != "" {
			f.currentScenario.StatusDetail += "\n"
		}
		f.currentScenario.StatusDetail += fmt.Sprintf("⋯ %s (pending)", step.Text)
	}
}

// NewOCSFFormatterWithParams creates a new OCSF formatter with test parameters
func NewOCSFFormatterWithParams(suite string, out io.Writer, params PortTestParams) formatters.Formatter {
	return &OCSFFormatter{
		out:      out,
		findings: make([]OCSFFinding, 0),
		params:   &params,
	}
}
