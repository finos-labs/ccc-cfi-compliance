package login

import (
	"fmt"
	"sync"
)

// Manager serializes login refresh. Whether to refresh is driven by each cloud's token expiry
// (JWT exp, Azure CLI expires_on, AWS_CREDENTIAL_EXPIRATION, etc.), not a separate wall-clock timer.
type Manager struct {
	mu sync.Mutex
}

// NewManager creates a Manager.
func NewManager() *Manager {
	return &Manager{}
}

// EnsureLoginToken runs the provider login path when the current token is missing, unreadable,
// or expires within refreshSkew.
// cloud must be "aws", "azure", or "gcp" (see factory.CloudProvider).
func (m *Manager) EnsureLoginToken(cloud string) error {
	if cloud == "" {
		return fmt.Errorf("login: empty cloud")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ok, err := cloudTokenOK(cloud)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	switch cloud {
	case "aws":
		err = refreshAWSCLI()
	case "azure":
		err = refreshAzureCLI()
	case "gcp":
		err = refreshGCPCLI()
	default:
		return fmt.Errorf("login: unsupported cloud %q", cloud)
	}
	return err
}
