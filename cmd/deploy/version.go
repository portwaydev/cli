package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

func hasUncommittedChanges(configDir string) bool {
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = configDir
	statusOutput, err := statusCmd.Output()
	return err == nil && len(statusOutput) > 0
}

func confirmUncommittedChanges() error {
	confirmed := false

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("There are uncommitted changes in your working directory. Continue?").
				Value(&confirmed),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run prompt: %w", err)
	}

	if !confirmed {
		fmt.Print("\nAborted due to uncommitted changes.\n\n")
		os.Exit(1)
	}

	return nil
}

func determineVersion() (string, error) {
	// Get git commit hash
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit hash: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}
