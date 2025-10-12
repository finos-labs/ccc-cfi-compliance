package example

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// Global function for godog CLI (required by godog)
// Uses the first entry from portTestMatrix for default parameters
func InitializeScenario(ctx *godog.ScenarioContext) {
	suite := cloud.NewTestSuite()
	suite.InitializeScenarioWithParams(ctx, portTestMatrix[0].params)
}

// Test matrix for different protocols over secure and plaintext ports
var portTestMatrix = []struct {
	name        string
	params      reporters.TestParams
	description string
}{
	// HTTPS - Secure
	{
		name: "HTTPS_Secure",
		params: reporters.TestParams{
			PortNumber:  "443",
			HostName:    "robmoff.at",
			Protocol:    "http",
			ServiceType: "web",
			Region:      "eu-west-1",
			Provider:    "aws",
			Labels:      []string{"https", "tls", "tlp-amber"},
			UID:         "https-443",
		},
		description: "HTTPS on port 443 (TLS)",
	},
	// HTTP - Plaintext
	{
		name: "HTTP_Plaintext",
		params: reporters.TestParams{
			PortNumber:  "80",
			HostName:    "robmoff.at",
			Protocol:    "http",
			ServiceType: "web",
			Region:      "eu-west-1",
			Provider:    "aws",
			Labels:      []string{"http", "plaintext", "tlp-amber"},
			UID:         "http-80",
		},
		description: "HTTP on port 80 (plaintext)",
	},
	// SSH - Secure
	{
		name: "SSH_Secure",
		params: reporters.TestParams{
			PortNumber:  "22",
			HostName:    "172.104.252.249", // automation.risk-first.org, change later.
			Protocol:    "ssh",
			ServiceType: "ssh",
			Region:      "eu-west-1",
			Provider:    "aws",
			Labels:      []string{"ssh", "encrypted", "tlp-amber"},
			UID:         "ssh-22",
		},
		description: "SSH on port 22 (encrypted)",
	},
	// SMTP - Secure (SMTPS)
	{
		name: "SMTPS_Secure",
		params: reporters.TestParams{
			PortNumber:  "465",
			HostName:    "secure.emailsrvr.com",
			Protocol:    "smtp",
			ServiceType: "mail",
			Region:      "eu-west-1",
			Provider:    "aws",
			Labels:      []string{"smtp", "tls", "tlp-amber"},
			UID:         "smtp-465",
		},
		description: "SMTPS on port 465 (TLS)",
	},
	// SMTP - Plaintext with STARTTLS
	{
		name: "SMTP_STARTTLS",
		params: reporters.TestParams{
			PortNumber:  "587",
			HostName:    "secure.emailsrvr.com",
			Protocol:    "smtp",
			ServiceType: "mail",
			Region:      "eu-west-1",
			Provider:    "aws",
			Labels:      []string{"smtp", "tls", "tlp-amber"},
			UID:         "smtp-587",
		},
		description: "SMTP on port 587 (STARTTLS)",
	},
	// // FTP - Plaintext
	// {
	// 	name: "FTP_Plaintext",
	// 	params: TestParams{
	// 		PortNumber:  "21",
	// 		HostName:    "robmoff.at",
	// 		Protocol:    "ftp",
	// 		ServiceType: "file",
	// 		Region:      "eu-west-1",
	// 		Provider:    "aws",
	// 		Labels:      []string{"ftp", "plaintext", "tlp-amber"},
	// 		UID:         "ftp-21",
	// 	},
	// 	description: "FTP on port 21 (plaintext)",
	// },
	// // FTPS - Secure
	// {
	// 	name: "FTPS_Secure",
	// 	params: TestParams{
	// 		PortNumber:  "990",
	// 		HostName:    "robmoff.at",
	// 		Protocol:    "ftp",
	// 		ServiceType: "file",
	// 		Region:      "eu-west-1",
	// 		Provider:    "aws",
	// 		Labels:      []string{"ftp", "tls", "tlp-red"},
	// 		UID:         "ftp-990",
	// 	},
	// 	description: "FTPS on port 990 (TLS)",
	// },
}

// TestCloudPortFeatures tests the C01 features with example parameters
func TestCloudPortFeatures(t *testing.T) {
	for _, tc := range portTestMatrix {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)
			reportPath := "output/port-test-" + tc.params.HostName + "-" + tc.params.PortNumber
			cloud.RunPortTests(t, tc.params, "../../../features/CCC.Core/CO1", reportPath)
		})
	}
}
