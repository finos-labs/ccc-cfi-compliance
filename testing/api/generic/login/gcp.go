package login

import (
	"fmt"
	"os"
	"os/exec"
)

func refreshGCPCLI() error {
	if keyFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); keyFile != "" {
		cmd := exec.Command("gcloud", "auth", "activate-service-account", "--key-file="+keyFile)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gcloud auth activate-service-account: %w: %s", err, string(out))
		}
		return nil
	}

	cmd := exec.Command("gcloud", "auth", "print-access-token", "--quiet")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gcloud auth print-access-token: %w: %s (set GOOGLE_APPLICATION_CREDENTIALS or run gcloud auth login locally)", err, string(out))
	}
	_ = out
	return nil
}
