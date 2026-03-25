package vpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

type cn03TrialMatrix struct {
	ReceiverVpcID             string
	PeerOwnerID               string
	AllowedRequesterVpcIDs    []string
	DisallowedRequesterVpcIDs []string
}

func (s *AWSVPCService) EvaluatePeerAgainstAllowList(peerVpcID string) (map[string]interface{}, error) {
	peerVpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", peerVpcID))
	if peerVpcIDStr == "" {
		return nil, fmt.Errorf("peerVpcID is required")
	}

	allowedIDs, source, err := s.resolveCN03AllowedRequesterVpcIDs()
	if err != nil {
		return nil, err
	}

	allowed := false
	for _, allowedID := range allowedIDs {
		if allowedID == peerVpcIDStr {
			allowed = true
			break
		}
	}

	reason := "CN03 allow-list is not defined via environment variables or trial matrix file; classification is non-enforcing until IAM/SCP guardrail is configured"
	if len(allowedIDs) > 0 {
		if allowed {
			reason = "requester VPC exists in CN03 allow-list; expected enforcement outcome is allow"
		} else {
			reason = "requester VPC does not exist in CN03 allow-list; expected enforcement outcome is deny"
		}
	}

	return map[string]interface{}{
		"PeerVpcId":              peerVpcIDStr,
		"Allowed":                allowed,
		"AllowedListDefined":     len(allowedIDs) > 0,
		"AllowedListCount":       len(allowedIDs),
		"AllowedRequesterVpcIds": allowedIDs,
		"AllowedPeerVpcIds":      allowedIDs,
		"AllowListSource":        source,
		"Reason":                 reason,
	}, nil
}

func (s *AWSVPCService) AttemptVpcPeeringDryRun(requesterVpcID, peerVpcID string) (map[string]interface{}, error) {
	return s.attemptVpcPeeringDryRunWithOwner(requesterVpcID, peerVpcID, cn03PeerOwnerID())
}

func (s *AWSVPCService) LoadVpcPeeringTrialMatrix(filePath string) (map[string]interface{}, error) {
	matrix, resolvedPath, err := s.loadCN03TrialMatrix(filePath)
	if err != nil {
		return nil, err
	}

	allRequesterIDs := make([]string, 0, len(matrix.AllowedRequesterVpcIDs)+len(matrix.DisallowedRequesterVpcIDs))
	allRequesterIDs = append(allRequesterIDs, matrix.AllowedRequesterVpcIDs...)
	allRequesterIDs = append(allRequesterIDs, matrix.DisallowedRequesterVpcIDs...)

	return map[string]interface{}{
		"FilePath":                   resolvedPath,
		"ReceiverVpcId":              matrix.ReceiverVpcID,
		"PeerOwnerId":                matrix.PeerOwnerID,
		"AllowedRequesterVpcIds":     matrix.AllowedRequesterVpcIDs,
		"DisallowedRequesterVpcIds":  matrix.DisallowedRequesterVpcIDs,
		"AllowedPeerVpcIds":          matrix.AllowedRequesterVpcIDs,
		"DisallowedPeerVpcIds":       matrix.DisallowedRequesterVpcIDs,
		"AllowedCount":               len(matrix.AllowedRequesterVpcIDs),
		"DisallowedCount":            len(matrix.DisallowedRequesterVpcIDs),
		"AllRequesterVpcIds":         allRequesterIDs,
		"AllRequesterCount":          len(allRequesterIDs),
		"AllowedListDefined":         len(matrix.AllowedRequesterVpcIDs) > 0,
		"DisallowedListDefined":      len(matrix.DisallowedRequesterVpcIDs) > 0,
		"ReceiverVpcIdMatchesRegion": matrix.ReceiverVpcID != "",
	}, nil
}

func (s *AWSVPCService) RunVpcPeeringDryRunTrialsFromFile(filePath string) (map[string]interface{}, error) {
	matrix, resolvedPath, err := s.loadCN03TrialMatrix(filePath)
	if err != nil {
		return nil, err
	}
	if matrix.ReceiverVpcID == "" {
		return nil, fmt.Errorf("CN03 trial matrix file %s is missing receiver_vpc_id", resolvedPath)
	}

	trials := make([]interface{}, 0, len(matrix.AllowedRequesterVpcIDs)+len(matrix.DisallowedRequesterVpcIDs))
	unexpectedCount := 0

	runTrials := func(requesterIDs []string, expectedAllowed bool) error {
		for _, requesterID := range requesterIDs {
			evidence, dryRunErr := s.attemptVpcPeeringDryRunWithOwner(requesterID, matrix.ReceiverVpcID, matrix.PeerOwnerID)
			if dryRunErr != nil {
				return dryRunErr
			}

			actualAllowed := boolFromEvidence(evidence["DryRunAllowed"])
			matchesExpectation := actualAllowed == expectedAllowed
			if !matchesExpectation {
				unexpectedCount++
			}

			trial := map[string]interface{}{
				"RequesterVpcId":     requesterID,
				"ReceiverVpcId":      matrix.ReceiverVpcID,
				"ExpectedAllowed":    expectedAllowed,
				"DryRunAllowed":      actualAllowed,
				"MatchesExpectation": matchesExpectation,
				"ExitCode":           evidence["ExitCode"],
				"ErrorCode":          evidence["ErrorCode"],
				"Stderr":             evidence["Stderr"],
			}

			trials = append(trials, trial)
		}
		return nil
	}

	if err := runTrials(matrix.DisallowedRequesterVpcIDs, false); err != nil {
		return nil, err
	}
	if err := runTrials(matrix.AllowedRequesterVpcIDs, true); err != nil {
		return nil, err
	}

	totalTrials := len(trials)
	return map[string]interface{}{
		"FilePath":                  resolvedPath,
		"ReceiverVpcId":             matrix.ReceiverVpcID,
		"PeerOwnerId":               matrix.PeerOwnerID,
		"AllowedRequesterVpcIds":    matrix.AllowedRequesterVpcIDs,
		"DisallowedRequesterVpcIds": matrix.DisallowedRequesterVpcIDs,
		"AllowedCount":              len(matrix.AllowedRequesterVpcIDs),
		"DisallowedCount":           len(matrix.DisallowedRequesterVpcIDs),
		"TotalTrials":               totalTrials,
		"UnexpectedCount":           unexpectedCount,
		"Compliant":                 totalTrials > 0 && unexpectedCount == 0,
		"Trials":                    trials,
	}, nil
}

func (s *AWSVPCService) attemptVpcPeeringDryRunWithOwner(requesterVpcID, peerVpcID, peerOwnerID string) (map[string]interface{}, error) {
	requesterVpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", requesterVpcID))
	peerVpcIDStr := strings.TrimSpace(fmt.Sprintf("%v", peerVpcID))
	peerOwnerIDStr := strings.TrimSpace(fmt.Sprintf("%v", peerOwnerID))

	if requesterVpcIDStr == "" {
		return nil, fmt.Errorf("requesterVpcID is required")
	}
	if peerVpcIDStr == "" {
		return nil, fmt.Errorf("peerVpcID is required")
	}

	input := &ec2.CreateVpcPeeringConnectionInput{
		VpcId:     aws.String(requesterVpcIDStr),
		PeerVpcId: aws.String(peerVpcIDStr),
		DryRun:    aws.Bool(true),
	}
	if peerOwnerIDStr != "" {
		input.PeerOwnerId = aws.String(peerOwnerIDStr)
	}

	evidence := map[string]interface{}{
		"RequesterVpcId": requesterVpcIDStr,
		"PeerVpcId":      peerVpcIDStr,
		"ReceiverVpcId":  peerVpcIDStr,
		"PeerOwnerId":    peerOwnerIDStr,
		"DryRunAllowed":  false,
		"ExitCode":       1,
		"ErrorCode":      "",
		"Stderr":         "",
		"Reason":         "request denied",
	}

	_, err := s.client.CreateVpcPeeringConnection(s.ctx, input)
	if err == nil {
		evidence["DryRunAllowed"] = true
		evidence["ExitCode"] = 0
		evidence["Reason"] = "dry-run call returned success; request would be allowed"
		return s.enrichCN03EnforcementEvidence(requesterVpcIDStr, evidence), nil
	}

	errText := strings.TrimSpace(err.Error())
	evidence["Stderr"] = errText
	evidence["Reason"] = errText

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		errorCode := strings.TrimSpace(apiErr.ErrorCode())
		evidence["ErrorCode"] = errorCode

		if strings.EqualFold(errorCode, "DryRunOperation") {
			evidence["DryRunAllowed"] = true
			evidence["ExitCode"] = 0
			evidence["Reason"] = "DryRunOperation indicates request would be allowed"
		}
		return s.enrichCN03EnforcementEvidence(requesterVpcIDStr, evidence), nil
	}

	if strings.Contains(strings.ToLower(errText), "dryrunoperation") {
		evidence["DryRunAllowed"] = true
		evidence["ExitCode"] = 0
		evidence["ErrorCode"] = "DryRunOperation"
		evidence["Reason"] = "dry-run response indicates request would be allowed"
	}

	return s.enrichCN03EnforcementEvidence(requesterVpcIDStr, evidence), nil
}

func (s *AWSVPCService) enrichCN03EnforcementEvidence(requesterVpcID string, evidence map[string]interface{}) map[string]interface{} {
	allowedIDs, source, err := s.resolveCN03AllowedRequesterVpcIDs()
	if err != nil {
		evidence["AllowListDefined"] = false
		evidence["AllowListSource"] = ""
		evidence["RequesterInAllowList"] = false
		evidence["GuardrailExpectation"] = ""
		evidence["GuardrailMismatch"] = false
		evidence["Reason"] = fmt.Sprintf("%v; CN03 allow-list resolution failed: %v", evidence["Reason"], err)
		return evidence
	}

	allowListDefined := len(allowedIDs) > 0
	requesterInAllowList := false
	for _, allowedID := range allowedIDs {
		if allowedID == requesterVpcID {
			requesterInAllowList = true
			break
		}
	}

	evidence["AllowListDefined"] = allowListDefined
	evidence["AllowListSource"] = source
	evidence["RequesterInAllowList"] = requesterInAllowList

	if !allowListDefined {
		evidence["GuardrailExpectation"] = ""
		evidence["GuardrailMismatch"] = false
		evidence["Reason"] = fmt.Sprintf("%v; CN03 allow-list is not defined, so enforcement expectation cannot be computed", evidence["Reason"])
		return evidence
	}

	expectedAllowed := requesterInAllowList
	guardrailExpectation := "deny"
	if expectedAllowed {
		guardrailExpectation = "allow"
	}

	actualAllowed := boolFromEvidence(evidence["DryRunAllowed"])
	// Keep ExitCode semantics stable for feature assertions:
	// denied dry-run should always report non-zero exit code.
	if !actualAllowed {
		switch v := evidence["ExitCode"].(type) {
		case int:
			if v <= 0 {
				evidence["ExitCode"] = 1
			}
		case int32:
			if v <= 0 {
				evidence["ExitCode"] = 1
			}
		case int64:
			if v <= 0 {
				evidence["ExitCode"] = 1
			}
		case float64:
			if v <= 0 {
				evidence["ExitCode"] = 1
			}
		case string:
			if strings.TrimSpace(v) == "" || strings.TrimSpace(v) == "0" {
				evidence["ExitCode"] = 1
			}
		default:
			evidence["ExitCode"] = 1
		}
	}
	guardrailMismatch := actualAllowed != expectedAllowed

	evidence["GuardrailExpectation"] = guardrailExpectation
	evidence["GuardrailMismatch"] = guardrailMismatch
	if guardrailMismatch {
		evidence["Reason"] = fmt.Sprintf("%v; CN03 guardrail mismatch: allow-list expects %s for requester %s", evidence["Reason"], guardrailExpectation, requesterVpcID)
	} else {
		evidence["Reason"] = fmt.Sprintf("%v; CN03 guardrail aligned: allow-list expects %s for requester %s", evidence["Reason"], guardrailExpectation, requesterVpcID)
	}

	return evidence
}

// resolveCN03AllowedRequesterVpcIDs collects the CN03 allow-list from all
// available sources and returns a deduplicated union. Sources:
//   - Dynamic: env vars set from cn03-feature.env CI artifacts
//     (CN03_ALLOWED_REQUESTER_VPC_IDS CSV, CN03_ALLOWED_REQUESTER_VPC_ID_1..N,
//     CN03_PEER_TRIAL_MATRIX_FILE, and legacy CN03_ALLOWED_PEER_VPC_* aliases)
//   - Manual: cn03-allowed-requester-vpc-ids in environment.yaml vpc service config
//
// If no source yields any IDs the caller should skip CN03 checks — no criteria
// can be established without at least one known allowed requester.
func (s *AWSVPCService) resolveCN03AllowedRequesterVpcIDs() ([]string, string, error) {
	var all []string
	var sources []string

	// Dynamic source: env vars (populated from cn03-feature.env CI artifacts)
	if ids := normalizeStringList([]string{os.Getenv("CN03_ALLOWED_REQUESTER_VPC_IDS")}); len(ids) > 0 {
		all = append(all, ids...)
		sources = append(sources, "CN03_ALLOWED_REQUESTER_VPC_IDS")
	}
	if ids := cn03IndexedEnvValues("CN03_ALLOWED_REQUESTER_VPC_ID_"); len(ids) > 0 {
		all = append(all, ids...)
		sources = append(sources, "CN03_ALLOWED_REQUESTER_VPC_ID_1..N")
	}
	if ids := normalizeStringList([]string{os.Getenv("CN03_ALLOWED_PEER_VPC_IDS")}); len(ids) > 0 {
		all = append(all, ids...)
		sources = append(sources, "CN03_ALLOWED_PEER_VPC_IDS")
	}
	if ids := cn03IndexedEnvValues("CN03_ALLOWED_PEER_VPC_ID_"); len(ids) > 0 {
		all = append(all, ids...)
		sources = append(sources, "CN03_ALLOWED_PEER_VPC_ID_1..N")
	}
	if matrixPath := strings.TrimSpace(os.Getenv("CN03_PEER_TRIAL_MATRIX_FILE")); matrixPath != "" {
		matrix, resolvedPath, err := s.loadCN03TrialMatrix(matrixPath)
		if err != nil {
			return nil, "", err
		}
		if len(matrix.AllowedRequesterVpcIDs) > 0 {
			all = append(all, matrix.AllowedRequesterVpcIDs...)
			sources = append(sources, fmt.Sprintf("CN03_PEER_TRIAL_MATRIX_FILE (%s)", resolvedPath))
		}
	}

	// Manual source: cn03-allowed-requester-vpc-ids from environment.yaml vpc service config
	if svcProps := s.instance.ServiceProperties("vpc"); svcProps != nil {
		if raw, ok := svcProps["cn03-allowed-requester-vpc-ids"]; ok {
			if ids := cn03StringSlice(raw); len(ids) > 0 {
				all = append(all, ids...)
				sources = append(sources, "environment.yaml/vpc/cn03-allowed-requester-vpc-ids")
			}
		}
	}

	combined := normalizeStringList(all)
	source := strings.Join(sources, ", ")
	if source == "" {
		source = "none"
	}
	return combined, source, nil
}

func (s *AWSVPCService) loadCN03TrialMatrix(filePath string) (cn03TrialMatrix, string, error) {
	resolvedPath := strings.TrimSpace(filePath)
	if resolvedPath == "" {
		resolvedPath = strings.TrimSpace(os.Getenv("CN03_PEER_TRIAL_MATRIX_FILE"))
	}
	if resolvedPath == "" {
		return cn03TrialMatrix{}, "", fmt.Errorf("filePath is required (or set CN03_PEER_TRIAL_MATRIX_FILE)")
	}

	if !filepath.IsAbs(resolvedPath) {
		if absPath, absErr := filepath.Abs(resolvedPath); absErr == nil {
			resolvedPath = absPath
		}
	}

	fileData, err := os.ReadFile(resolvedPath)
	if err != nil {
		return cn03TrialMatrix{}, "", fmt.Errorf("failed to read CN03 trial matrix file %s: %w", resolvedPath, err)
	}

	raw := make(map[string]interface{})
	if err := json.Unmarshal(fileData, &raw); err != nil {
		return cn03TrialMatrix{}, "", fmt.Errorf("failed to parse CN03 trial matrix file %s: %w", resolvedPath, err)
	}

	// Multiple field name aliases are accepted for format flexibility across
	// manually authored files and export-cn03-artifacts.sh generated outputs.
	receiverVpcID := firstNonEmptyString(
		cn03String(raw["receiver_vpc_id"]),
		cn03String(raw["peer_vpc_id"]),
		cn03String(raw["target_vpc_id"]),
		cn03String(raw["receiverVpcId"]),
		cn03String(raw["peerVpcId"]),
		cn03String(raw["targetVpcId"]),
	)

	allowedRequesterIDs := make([]string, 0)
	allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(raw["allowed_requester_vpc_ids"])...)
	allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(raw["allowed_peer_vpc_ids"])...)
	allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(raw["allowed_vpc_ids"])...)

	disallowedRequesterIDs := make([]string, 0)
	disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(raw["disallowed_requester_vpc_ids"])...)
	disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(raw["disallowed_peer_vpc_ids"])...)
	disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(raw["disallowed_vpc_ids"])...)

	if requesters, ok := raw["requesters"].(map[string]interface{}); ok {
		allowedRequesterIDs = append(allowedRequesterIDs, cn03StringSlice(requesters["allowed"])...)
		disallowedRequesterIDs = append(disallowedRequesterIDs, cn03StringSlice(requesters["disallowed"])...)
		receiverVpcID = firstNonEmptyString(
			receiverVpcID,
			cn03String(requesters["receiver_vpc_id"]),
			cn03String(requesters["receiverVpcId"]),
		)
	}

	allowedRequesterIDs = normalizeStringList(allowedRequesterIDs)
	disallowedRequesterIDs = normalizeStringList(disallowedRequesterIDs)
	if len(allowedRequesterIDs) == 0 && len(disallowedRequesterIDs) == 0 {
		return cn03TrialMatrix{}, "", fmt.Errorf("CN03 trial matrix file %s does not define any requester VPC IDs", resolvedPath)
	}

	peerOwnerID := firstNonEmptyString(
		cn03String(raw["peer_owner_id"]),
		cn03String(raw["peerOwnerId"]),
		cn03PeerOwnerID(),
	)

	return cn03TrialMatrix{
		ReceiverVpcID:             receiverVpcID,
		PeerOwnerID:               peerOwnerID,
		AllowedRequesterVpcIDs:    allowedRequesterIDs,
		DisallowedRequesterVpcIDs: disallowedRequesterIDs,
	}, resolvedPath, nil
}

// cn03IndexedEnvValues reads up to 99 sequentially numbered env vars with the
// given prefix (e.g. CN03_ALLOWED_REQUESTER_VPC_ID_1..99). Iteration stops at
// the first gap but collects all non-empty values regardless of order.
func cn03IndexedEnvValues(prefix string) []string {
	values := make([]string, 0)
	for i := 1; i <= 99; i++ {
		envKey := fmt.Sprintf("%s%d", prefix, i)
		rawValue := strings.TrimSpace(os.Getenv(envKey))
		if rawValue == "" {
			continue
		}
		values = append(values, rawValue)
	}
	return normalizeStringList(values)
}

func cn03String(value interface{}) string {
	if value == nil {
		return ""
	}
	out := strings.TrimSpace(fmt.Sprintf("%v", value))
	if out == "<nil>" {
		return ""
	}
	return out
}

func cn03StringSlice(value interface{}) []string {
	switch typedValue := value.(type) {
	case nil:
		return []string{}
	case string:
		return normalizeStringList([]string{typedValue})
	case []string:
		return normalizeStringList(typedValue)
	case []interface{}:
		items := make([]string, 0, len(typedValue))
		for _, item := range typedValue {
			items = append(items, cn03String(item))
		}
		return normalizeStringList(items)
	default:
		return normalizeStringList([]string{fmt.Sprintf("%v", typedValue)})
	}
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, rawValue := range values {
		for _, item := range strings.Split(rawValue, ",") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			if _, exists := seen[trimmed]; exists {
				continue
			}
			seen[trimmed] = struct{}{}
			normalized = append(normalized, trimmed)
		}
	}

	return normalized
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && trimmed != "<nil>" {
			return trimmed
		}
	}
	return ""
}

func cn03PeerOwnerID() string {
	for _, key := range []string{"CN03_PEER_OWNER_ID", "PEER_OWNER_ID"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}
