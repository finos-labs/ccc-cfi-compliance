package login

import (
	"fmt"
	"os/exec"
)

func refreshAWSCLI() error {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("aws sts get-caller-identity (refresh credentials / IAM role): %w: %s", err, string(out))
	}
	return nil
}
