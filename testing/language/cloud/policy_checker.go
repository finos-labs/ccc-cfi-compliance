package cloud

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic/login"
	"github.com/finos-labs/ccc-cfi-compliance/testing/types"
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
func (c *PolicyChecker) LoadPolicy(policyPath string) (*types.PolicyDefinition, error) {
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file %s: %w", policyPath, err)
	}

	var policy types.PolicyDefinition
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy file %s: %w", policyPath, err)
	}

	return &policy, nil
}

// SubstituteParams replaces parameter placeholders in a query string using values from Props
// Returns an error if any placeholder references an unknown parameter
func (c *PolicyChecker) SubstituteParams(query string, props map[string]interface{}) (string, error) {
	// Find all ${...} placeholders in the query
	placeholderRegex := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := placeholderRegex.FindAllStringSubmatch(query, -1)

	// Check that all placeholders have corresponding values
	var missingParams []string
	for _, match := range matches {
		paramName := match[1]
		if _, exists := props[paramName]; !exists {
			missingParams = append(missingParams, paramName)
		}
	}

	if len(missingParams) > 0 {
		return "", fmt.Errorf("unknown parameter(s) in policy query: %v (available: %v)",
			missingParams, getMapKeys(props))
	}

	// Replace all placeholders with their values
	result := placeholderRegex.ReplaceAllStringFunc(query, func(placeholder string) string {
		paramName := placeholder[2 : len(placeholder)-1] // strip ${ and }
		return fmt.Sprintf("%v", props[paramName])
	})

	return result, nil
}

// getMapKeys returns the keys of a map as a slice
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// InferCloudFromPolicyQuery returns which cloud a policy shell query targets ("aws", "azure", "gcp"), if recognizable.
func InferCloudFromPolicyQuery(query string) (string, bool) {
	s := strings.TrimSpace(strings.ReplaceAll(query, "\\\n", " "))
	switch {
	case strings.HasPrefix(strings.ToLower(s), "az "):
		return "azure", true
	case strings.HasPrefix(strings.ToLower(s), "aws "):
		return "aws", true
	case strings.HasPrefix(strings.ToLower(s), "gcloud "):
		return "gcp", true
	default:
		return "", false
	}
}

// ExecuteQuery runs a shell query and returns the output
func (c *PolicyChecker) ExecuteQuery(query string) (string, error) {
	// Clean up the query (remove line continuations, extra whitespace)
	cleanQuery := strings.ReplaceAll(query, "\\\n", " ")
	cleanQuery = strings.TrimSpace(cleanQuery)

	if cloud, ok := InferCloudFromPolicyQuery(cleanQuery); ok {
		if err := login.Default.EnsureLoginToken(cloud); err != nil {
			return "", fmt.Errorf("login: %w", err)
		}
	}

	cmd := exec.Command("sh", "-c", cleanQuery)
	cmd.Env = login.EnvForPolicyQuery(cleanQuery)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("query execution failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// EvaluateRule checks if a single rule passes against the query output.
// If validation_rule is present, regex match is used. Otherwise, actual values
// must be a subset of expected_values (allowlist check).
func (c *PolicyChecker) EvaluateRule(rule types.Rule, queryOutput string, props map[string]interface{}) types.RuleResult {
	result := types.RuleResult{
		JSONPath:       rule.JSONPath,
		ExpectedValues: make([]string, 0),
		ValidationRule: rule.ValidationRule,
		Description:    rule.Description,
		Passed:         false,
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

	// Collect actual values - support single value or array
	var actualValues []string
	if value == nil {
		actualValues = []string{"null"}
	} else {
		actualValues = valuesToSlice(value)
	}
	result.ActualValue = fmt.Sprintf("%v", actualValues)

	// Resolve expected values (allowlist) - may reference props e.g. ${PermittedAccountIds}
	allowedSet := resolveExpectedValues(rule.ExpectedValues, props)
	for k := range allowedSet {
		result.ExpectedValues = append(result.ExpectedValues, k)
	}

	// Check against validation rule (regex) or expected values (allowlist)
	if rule.ValidationRule != "" {
		regex, err := regexp.Compile(rule.ValidationRule)
		if err != nil {
			result.Error = fmt.Sprintf("invalid validation regex %s: %v", rule.ValidationRule, err)
			return result
		}
		// Match against actual values, not the slice representation (e.g. "[value]")
		for _, av := range actualValues {
			if regex.MatchString(av) {
				result.Passed = true
				break
			}
		}
	} else {
		// Allowlist: every actual value must be in expected set. Empty actual = pass.
		result.Passed = true
		for _, a := range actualValues {
			if a != "" && !allowedSet[a] {
				result.Passed = false
				break
			}
		}
	}

	return result
}

// valuesToSlice converts a jsonpath result to a slice of strings.
func valuesToSlice(value interface{}) []string {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if item != nil {
				out = append(out, strings.TrimSpace(fmt.Sprintf("%v", item)))
			}
		}
		return out
	default:
		return []string{strings.TrimSpace(fmt.Sprintf("%v", value))}
	}
}

// resolveExpectedValues builds the allowlist from ExpectedValues, resolving ${Param} refs from props.
func resolveExpectedValues(expected []any, props map[string]interface{}) map[string]bool {
	allowed := make(map[string]bool)
	if props == nil {
		return allowed
	}
	paramRef := regexp.MustCompile(`^\$\{([^}]+)\}$`)
	for _, ev := range expected {
		s := strings.TrimSpace(fmt.Sprintf("%v", ev))
		if m := paramRef.FindStringSubmatch(s); len(m) == 2 {
			paramName := m[1]
			if pv, ok := props[paramName]; ok && pv != nil {
				// Param value can be comma-separated string or slice
				switch v := pv.(type) {
				case []interface{}:
					for _, item := range v {
						allowed[strings.TrimSpace(fmt.Sprintf("%v", item))] = true
					}
			default:
				raw := fmt.Sprintf("%v", pv)
				// Handle comma-separated string or Go slice format "[a b c]"
				if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
					raw = strings.Trim(raw, "[]")
					for _, part := range strings.Fields(raw) {
						if t := strings.TrimSpace(part); t != "" {
							allowed[t] = true
						}
					}
				} else {
					for _, part := range strings.Split(raw, ",") {
						if t := strings.TrimSpace(part); t != "" {
							allowed[t] = true
						}
					}
				}
			}
			}
		} else if s != "" {
			allowed[s] = true
		}
	}
	return allowed
}

// RunPolicy executes a complete policy check using values from Props
func (c *PolicyChecker) RunPolicy(props map[string]interface{}, policyPath string) (*types.PolicyResult, error) {
	// Load the policy
	policyDef, err := c.LoadPolicy(policyPath)
	if err != nil {
		return nil, err
	}

	result := &types.PolicyResult{
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
	queryExecuted, err := c.SubstituteParams(policyDef.Query, props)
	if err != nil {
		result.QueryError = err.Error()
		result.Passed = false
		return result, nil
	}
	result.QueryExecuted = queryExecuted

	// Execute the query
	output, err := c.ExecuteQuery(result.QueryExecuted)
	result.QueryOutput = output
	if err != nil {
		result.QueryError = err.Error()
		result.Passed = false
		return result, nil // Return result with error, don't fail completely
	}

	// Evaluate each rule
	result.RuleResults = make([]types.RuleResult, len(policyDef.Rules))
	for i, rule := range policyDef.Rules {
		ruleResult := c.EvaluateRule(rule, output, props)
		result.RuleResults[i] = ruleResult
		if !ruleResult.Passed {
			result.Passed = false
		}
	}

	return result, nil
}
