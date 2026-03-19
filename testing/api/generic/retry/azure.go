package retry

import (
	"errors"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// Default propagation retry parameters for Azure (RBAC and Graph API can take up to 5 min)
const (
	DefaultPropagationAttempts = 5
	DefaultPropagationDelay   = 60 * time.Second
)

// IsAzureRBACPropagationError returns true for 403 AuthorizationPermissionMismatch,
// which commonly occurs when RBAC role assignments have not yet propagated (up to 5 min).
func IsAzureRBACPropagationError(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.StatusCode == 403 && respErr.ErrorCode == "AuthorizationPermissionMismatch"
	}
	// Fallback: error may be wrapped or in string form (e.g. from policy/CLI output)
	if err != nil {
		msg := err.Error()
		return strings.Contains(msg, "403") && strings.Contains(msg, "AuthorizationPermissionMismatch")
	}
	return false
}

// IsAzureCredentialPropagationError returns true for AAD errors indicating
// the client secret has not yet propagated (typically within ~60s).
func IsAzureCredentialPropagationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "AADSTS7000215") ||
		strings.Contains(msg, "invalid_client") ||
		strings.Contains(msg, "unauthorized_client")
}

// IsAzureGraphAuthorizationDeniedError returns true for Microsoft Graph API
// 403 Authorization_RequestDenied, which can occur when Graph API permissions
// have not yet propagated after being granted (similar to RBAC propagation).
func IsAzureGraphAuthorizationDeniedError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "403") && strings.Contains(msg, "Authorization_RequestDenied")
}
