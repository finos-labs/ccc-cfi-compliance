package login

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Azure login refresh for policy checks and the test runner.
//
// GitHub Actions (cfi-test.yml) authenticates with azure/login@v1 using only:
//   - client-id  → AZURE_CLIENT_ID
//   - tenant-id  → AZURE_TENANT_ID
//   - subscription-id → AZURE_SUBSCRIPTION_ID
// and permissions.id-token: write (ACTIONS_ID_TOKEN_REQUEST_*). subscription-id is also passed
// there and exported as AZURE_SUBSCRIPTION_ID; we run az account set when it is set.
// Mid-job we repeat the same federated az login so CLI tokens stay valid.
//
// Locally (not GHA), falls back to interactive az login and AZURE_SUBSCRIPTION_ID if set.
func refreshAzureCLI() error {
	if inGitHubActionsWithAzureOIDC() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		return runAzLoginWithFederatedToken(ctx)
	}
	return runAzLoginInteractive()
}

func inGitHubActionsWithAzureOIDC() bool {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return false
	}
	if strings.TrimSpace(os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")) == "" ||
		os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN") == "" {
		return false
	}
	return os.Getenv("AZURE_CLIENT_ID") != "" && os.Getenv("AZURE_TENANT_ID") != ""
}

// runAzLoginWithFederatedToken matches azure/login@v1 OIDC: exchange GitHub ID token for az session.
func runAzLoginWithFederatedToken(ctx context.Context) error {
	reqURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	reqTok := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tenantID := os.Getenv("AZURE_TENANT_ID")

	u, err := url.Parse(reqURL)
	if err != nil {
		return fmt.Errorf("parse ACTIONS_ID_TOKEN_REQUEST_URL: %w", err)
	}
	q := u.Query()
	// Default matches Entra federated identity for GitHub unless the workflow sets audience on azure/login.
	q.Set("audience", firstNonEmpty(os.Getenv("AZURE_GITHUB_OIDC_AUDIENCE"), "api://AzureADTokenExchange"))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("github oidc request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+reqTok)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("github oidc token: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read oidc response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github oidc http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.Value == "" {
		return fmt.Errorf("parse oidc json: %w", err)
	}

	env := sanitizeAzureFedTokenEnv(os.Environ())
	login := exec.CommandContext(ctx, "az", "login", "--service-principal",
		"--username", clientID,
		"--tenant", tenantID,
		"--federated-token", payload.Value,
	)
	login.Env = env
	if out, err := login.CombinedOutput(); err != nil {
		return fmt.Errorf("az login (federated): %w: %s", err, string(out))
	}

	if sub := os.Getenv("AZURE_SUBSCRIPTION_ID"); sub != "" {
		set := exec.CommandContext(ctx, "az", "account", "set", "--subscription", sub)
		set.Env = env
		if out, err := set.CombinedOutput(); err != nil {
			return fmt.Errorf("az account set: %w: %s", err, string(out))
		}
	}
	return nil
}

func runAzLoginInteractive() error {
	cmd := exec.Command("az", "login")
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("az login: %w: %s", err, string(out))
	}
	if sub := os.Getenv("AZURE_SUBSCRIPTION_ID"); sub != "" {
		set := exec.Command("az", "account", "set", "--subscription", sub)
		set.Env = os.Environ()
		if out, err := set.CombinedOutput(); err != nil {
			return fmt.Errorf("az account set: %w: %s", err, string(out))
		}
	}
	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func sanitizeAzureFedTokenEnv(environ []string) []string {
	out := make([]string, 0, len(environ))
	for _, e := range environ {
		if strings.HasPrefix(e, "AZURE_FEDERATED_TOKEN=") || strings.HasPrefix(e, "AZURE_FEDERATED_TOKEN_FILE=") {
			continue
		}
		out = append(out, e)
	}
	return out
}
