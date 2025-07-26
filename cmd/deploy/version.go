package deploy

import (
	"cli/pkg/util"
	"crypto/md5"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
)

func determineVersion(composeFile []byte) (string, error) {
	// Check for pending changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err == nil && len(statusOutput) > 0 {
		if !util.IsCI() {
			response, err := pterm.DefaultInteractiveTextInput.Show("There are uncommitted changes. Continue? [y/N]")
			if err != nil {
				return "", fmt.Errorf("failed to get input: %w", err)
			}
			if response != "y" && response != "Y" {
				return "", fmt.Errorf("aborted due to uncommitted changes")
			}
		} else {
			pterm.Printf("%s  Warning: There are uncommitted changes\n", pterm.Yellow("⚠️"))
		}
	}

	// Get git commit hash
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Git command failed, fallback to MD5 hash of file
		hash := md5.Sum(composeFile)
		version := fmt.Sprintf("%x", hash)[:8] // Use first 8 chars of MD5 hash
		return version, nil
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}
