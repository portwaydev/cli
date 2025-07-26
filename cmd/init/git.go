package init

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func getRepoName() string {
	gitCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := gitCmd.Output()
	if err != nil {
		return ""
	}

	repoURL := strings.TrimSpace(string(output))
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")

	return repoName
}
