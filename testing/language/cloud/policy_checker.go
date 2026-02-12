package cloud

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
	"gopkg.in/yaml.v3"
)

// PolicyChecker handles loading and running policy checks
type PolicyChecker struct {
	// Base directory for policy files
	PolicyBaseDir string
}

// NewPolicyChecker creates a new policy checker
func NewPolicyChecker(baseDir string) *PolicyChecker {
	return &PolicyChecker{
		PolicyBaseDir: baseDir,
	}
}

// LoadPolicy loads a policy definition from a YAML file
func (c *PolicyChecker) LoadPolicy(policyPath string) (*environment.PolicyDefinition, error) {
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file %s: %w", policyPath, err)
	}

	var policy environment.PolicyDefinition
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy file %s: %w", policyPath, err)
	}

	return &policy, nil
}

// SubstituteParams replaces parameter placeholders in a query string
func (c *PolicyChecker) SubstituteParams(query string, params environment.TestParams) string {
	result := query

	// Map TestParams fields
	result = strings.ReplaceAll(result, "${ResourceName}", params.ResourceName)
	result = strings.ReplaceAll(result, "${UID}", params.UID)
	result = strings.ReplaceAll(result, "${ServiceType}", params.ServiceType)

	// Map CloudParams fields
	result = strings.ReplaceAll(result, "${Provider}", params.CloudParams.Provider)
	result = strings.ReplaceAll(result, "${Region}", params.CloudParams.Region)
	result = strings.ReplaceAll(result, "${AzureResourceGroup}", params.CloudParams.AzureResourceGroup)
	result = strings.ReplaceAll(result, "${AzureSubscriptionID}", params.CloudParams.AzureSubscriptionID)
	result = strings.ReplaceAll(result, "${AzureStorageAccount}", params.CloudParams.AzureStorageAccount)
	result = strings.ReplaceAll(result, "${GCPProjectID}", params.CloudParams.GCPProjectID)

	// Legacy parameter mappings (for backwards compatibility)
	result = strings.ReplaceAll(result, "${BUCKET_NAME}", params.ResourceName)
	result = strings.ReplaceAll(result, "${VPC_ID}", params.UID)
	result = strings.ReplaceAll(result, "${SECURITY_GROUP_ID}", params.UID)
	result = strings.ReplaceAll(result, "${LOAD_BALANCER_ARN}", params.UID)
	result = strings.ReplaceAll(result, "${LISTENER_ARN}", params.UID)
	result = strings.ReplaceAll(result, "${TRUST_STORE_ARN}", params.UID)
	result = strings.ReplaceAll(result, "${KMS_KEY_ID}", params.UID)

	return result
}

// ExecuteQuery runs a shell query and returns the output
func (c *PolicyChecker) ExecuteQuery(query string) (string, error) {
	// Clean up the query (remove line continuations, extra whitespace)
	cleanQuery := strings.ReplaceAll(query, "\\\n", " ")
	cleanQuery = strings.TrimSpace(cleanQuery)

	// Execute via shell
	cmd := exec.Command("sh", "-c", cleanQuery)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("query execution failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// EvaluateRule checks if a single rule passes against the query output
func (c *PolicyChecker) EvaluateRule(rule environment.Rule, queryOutput string) environment.RuleResult {
	result := environment.RuleResult{
		JSONPath:       rule.JSONPath,
		ExpectedValues: make([]string, len(rule.ExpectedValues)),
		ValidationRule: rule.ValidationRule,
		Description:    rule.Description,
		Passed:         false,
	}

	// Convert expected values to strings
	for i, v := range rule.ExpectedValues {
		result.ExpectedValues[i] = fmt.Sprintf("%v", v)
	}

	// Parse the JSON output
	var jsonData interface{}
	if err := json.Unmarshal([]byte(queryOutput), &jsonData); err != nil {
		result.Error = fmt.Sprintf("failed to parse JSON output: %v", err)
		return result
	}

	// Execute JSONPath query using PaesslerAG/jsonpath
	value, err := jsonpath.Get(rule.JSONPath, jsonData)
	if err != nil {
		result.Error = fmt.Sprintf("JSONPath query failed %s: %v", rule.JSONPath, err)
		return result
	}

	// Get the actual value - handle nil (JSON null) as "null" string for matching
	var actualValue string
	if value == nil {
		actualValue = "null"
	} else {
		actualValue = fmt.Sprintf("%v", value)
	}
	result.ActualValue = actualValue

	// Check against validation rule (regex)
	if rule.ValidationRule != "" {
		regex, err := regexp.Compile(rule.ValidationRule)
		if err != nil {
			result.Error = fmt.Sprintf("invalid validation regex %s: %v", rule.ValidationRule, err)
			return result
		}
		result.Passed = regex.MatchString(actualValue)
	} else {
		// Check against expected values
		for _, expected := range result.ExpectedValues {
			if actualValue == expected {
				result.Passed = true
				break
			}
		}
	}

	return result
}

// RunPolicy executes a complete policy check
func (c *PolicyChecker) RunPolicy(params environment.TestParams, policyPath string) (*environment.PolicyResult, error) {
	// Load the policy
	policyDef, err := c.LoadPolicy(policyPath)
	if err != nil {
		return nil, err
	}

	result := &environment.PolicyResult{
		PolicyPath:      policyPath,
		Name:            policyDef.Name,
		ServiceType:     policyDef.ServiceType,
		RequirementText: policyDef.RequirementText,
		ValidityScore:   policyDef.ValidityScore,
		ValidityComment: policyDef.ValidityCommentary,
		QueryTemplate:   policyDef.Query,
		Passed:          true, // Will be set to false if any rule fails
	}

	// Substitute parameters in the query
	result.QueryExecuted = c.SubstituteParams(policyDef.Query, params)

	// Execute the query
	output, err := c.ExecuteQuery(result.QueryExecuted)
	result.QueryOutput = output
	if err != nil {
		result.QueryError = err.Error()
		result.Passed = false
		return result, nil // Return result with error, don't fail completely
	}

	// Evaluate each rule
	result.RuleResults = make([]environment.RuleResult, len(policyDef.Rules))
	for i, rule := range policyDef.Rules {
		ruleResult := c.EvaluateRule(rule, output)
		result.RuleResults[i] = ruleResult
		if !ruleResult.Passed {
			result.Passed = false
		}
	}

	return result, nil
}
