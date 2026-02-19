package environment

// TestParams holds the parameters for port / service testing
// This is the single shared structure used by both cloud api and reporters
type TestParams struct {
	PortNumber          string      // Leave blank if not applicable (e.g., for services without specific ports)
	HostName            string      // Hostname or endpoint
	Protocol            string      // Protocol (e.g., "tcp", "https")
	ServiceType         string      // Type of service (e.g., "s3", "rds", "storage") - DEPRECATED, use ProviderServiceType
	ProviderServiceType string      // Cloud provider-specific service type (e.g., "s3", "rds", "Microsoft.Storage/storageAccounts")
	CatalogTypes        []string    // CCC catalog types to test with (e.g., "CCC.ObjStor", "CCC.RDMS", "CCC.VM", "CCC.Core")
	TagFilter           []string    // Tag filters to AND together (e.g., ["@CCC.Core", "@CCC.ObjStor"])
	Labels              []string    // Tags/labels from the resource
	UID                 string      // Unique identifier (ARN, resource ID, etc.)
	ResourceName        string      // Human-readable resource name extracted from ARN or resource ID
	ReportFile          string      // Base filename for output report (without extension), e.g., "bucket-name-service"
	ReportTitle         string      // Human-readable title for reports, e.g., "my-bucket" or "my-bucket.s3.us-east-1.amazonaws.com:443"
	CloudParams         CloudParams // details of the cloud environment
}

type CloudParams struct {
	Provider            string // Cloud provider ("aws", "azure", "gcp")
	Region              string // Cloud region
	AzureResourceGroup  string // Azure resource group name
	AzureSubscriptionID string // Azure subscription ID
	AzureStorageAccount string // Azure storage account name
	GCPProjectID        string // GCP Project ID
}

// ServiceTypes contains all known service types
var ServiceTypes = []string{
	"object-storage",
	"block-storage",
	"relational-database",
	"iam",
	"load-balancer",
	"security-group",
	"vpc",
}

// PolicyResult contains the complete result of a policy check
type PolicyResult struct {
	// Policy metadata
	PolicyPath      string `json:"policy_path" yaml:"policy_path"`
	Name            string `json:"name" yaml:"name"`
	ServiceType     string `json:"service_type" yaml:"service_type"`
	RequirementText string `json:"requirement_text" yaml:"requirement_text"`
	ValidityScore   int    `json:"validity_score" yaml:"validity_score"`
	ValidityComment string `json:"validity_commentary" yaml:"validity_commentary"`

	// Query execution
	QueryTemplate string `json:"query_template" yaml:"query_template"`
	QueryExecuted string `json:"query_executed" yaml:"query_executed"`
	QueryOutput   string `json:"query_output" yaml:"query_output"`
	QueryError    string `json:"query_error,omitempty" yaml:"query_error,omitempty"`

	// Overall result
	Passed bool `json:"passed" yaml:"passed"`

	// Individual rule results
	RuleResults []RuleResult `json:"rule_results" yaml:"rule_results"`
}

// RuleResult contains the result of evaluating a single rule
type RuleResult struct {
	JSONPath       string   `json:"jsonpath" yaml:"jsonpath"`
	ExpectedValues []string `json:"expected_values" yaml:"expected_values"`
	ValidationRule string   `json:"validation_rule" yaml:"validation_rule"`
	Description    string   `json:"description" yaml:"description"`

	// Evaluation results
	ActualValue string `json:"actual_value" yaml:"actual_value"`
	Passed      bool   `json:"passed" yaml:"passed"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// PolicyDefinition represents the structure of a policy YAML file
type PolicyDefinition struct {
	Name               string `yaml:"name"`
	ServiceType        string `yaml:"service_type"`
	RequirementText    string `yaml:"requirement_text"`
	ValidityScore      int    `yaml:"validity_score"`
	ValidityCommentary string `yaml:"validity_commentary"`
	Query              string `yaml:"query"`
	Rules              []Rule `yaml:"rules"`
}

// Rule represents a single validation rule in a policy
type Rule struct {
	JSONPath       string `yaml:"jsonpath"`
	ExpectedValues []any  `yaml:"expected_values"`
	ValidationRule string `yaml:"validation_rule"`
	Description    string `yaml:"description"`
	Todo           string `yaml:"todo,omitempty"`
}

// Parameter names that can be used in policy queries
// These map to fields in TestParams and CloudParams
const (
	// From TestParams
	ParamResourceName = "${ResourceName}" // Human-readable resource name
	ParamUID          = "${UID}"          // ARN, resource ID, etc.
	ParamServiceType  = "${ServiceType}"  // Service type (e.g., "object-storage")

	// From CloudParams
	ParamProvider            = "${Provider}"            // Cloud provider
	ParamRegion              = "${Region}"              // Cloud region
	ParamAzureResourceGroup  = "${AzureResourceGroup}"  // Azure resource group
	ParamAzureSubscriptionID = "${AzureSubscriptionID}" // Azure subscription ID
	ParamAzureStorageAccount = "${AzureStorageAccount}" // Azure storage account
	ParamGCPProjectID        = "${GCPProjectID}"        // GCP project ID

	// Legacy parameter names (for backwards compatibility during migration)
	ParamBucketName = "${BUCKET_NAME}" // Maps to ResourceName for object-storage
)
