package retry

import (
	"errors"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// IsAzureRBACPropagationError returns true for 403 AuthorizationPermissionMismatch,
// which commonly occurs when RBAC role assignments have not yet propagated (up to 5 min).
func IsAzureRBACPropagationError(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.StatusCode == 403 && respErr.ErrorCode == "AuthorizationPermissionMismatch"
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
