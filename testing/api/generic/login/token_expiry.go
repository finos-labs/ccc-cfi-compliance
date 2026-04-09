package login

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// refreshSkew is how much lifetime we want left before skipping a re-login.
const refreshSkew = 5 * time.Minute

func tokenFreshEnough(exp time.Time, ok bool) bool {
	if !ok || exp.IsZero() {
		return false
	}
	return time.Until(exp) > refreshSkew
}

// --- Azure ---

func azureManagementTokenExpiry() (time.Time, bool) {
	cmd := exec.Command("az", "account", "get-access-token",
		"--resource", "https://management.azure.com/",
		"-o", "json",
	)
	cmd.Env = sanitizeAzureFedTokenEnv(os.Environ())
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, false
	}
	var resp struct {
		AccessToken string `json:"accessToken"`
		ExpiresOn   int64  `json:"expires_on"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return time.Time{}, false
	}
	if resp.ExpiresOn > 0 {
		return time.Unix(resp.ExpiresOn, 0), true
	}
	return jwtExpiry(resp.AccessToken)
}

func azureTokenOK() bool {
	exp, ok := azureManagementTokenExpiry()
	return tokenFreshEnough(exp, ok)
}

// --- AWS ---

func awsSessionExpiry() (time.Time, bool) {
	raw := strings.TrimSpace(os.Getenv("AWS_CREDENTIAL_EXPIRATION"))
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true
	}
	// Seen in some tooling without colon in offset
	if t, err := time.Parse("2006-01-02T15:04:05Z0700", raw); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func awsTokenOK() bool {
	if exp, ok := awsSessionExpiry(); ok {
		return tokenFreshEnough(exp, true)
	}
	// No embedded expiry (long-lived keys): treat as OK if STS works.
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// --- GCP ---

func gcpAccessTokenExpiry() (time.Time, bool) {
	cmd := exec.Command("gcloud", "auth", "print-access-token", "--quiet")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, false
	}
	return jwtExpiry(strings.TrimSpace(string(out)))
}

func gcpTokenOK() bool {
	exp, ok := gcpAccessTokenExpiry()
	if ok {
		return tokenFreshEnough(exp, true)
	}
	return false
}

// cloudTokenOK returns true if the default credentials for that cloud look valid long enough.
func cloudTokenOK(cloud string) (bool, error) {
	switch cloud {
	case "azure":
		return azureTokenOK(), nil
	case "aws":
		return awsTokenOK(), nil
	case "gcp":
		return gcpTokenOK(), nil
	default:
		return false, fmt.Errorf("login: unsupported cloud %q", cloud)
	}
}
