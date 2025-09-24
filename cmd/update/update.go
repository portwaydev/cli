package update

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const (
	repoOwner = "portwaydev"
	repoName  = "cli"
)

func getLatestTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	req, _ := http.NewRequest("GET", url, nil)
	// Basic rate-limit friendly UA
	req.Header.Set("User-Agent", "portway-cli-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// very small parse to find tag_name without importing a JSON lib
	idx := bytes.Index(body, []byte("\"tag_name\""))
	if idx == -1 {
		return "", errors.New("tag_name not found in GitHub response")
	}
	// find first quote after colon
	colon := bytes.IndexByte(body[idx:], ':')
	if colon == -1 {
		return "", errors.New("invalid GitHub response format")
	}
	rest := body[idx+colon+1:]
	firstQuote := bytes.IndexByte(rest, '"')
	if firstQuote == -1 {
		return "", errors.New("invalid GitHub response format")
	}
	rest = rest[firstQuote+1:]
	secondQuote := bytes.IndexByte(rest, '"')
	if secondQuote == -1 {
		return "", errors.New("invalid GitHub response format")
	}
	return string(rest[:secondQuote]), nil
}

func assetName() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		// ok
	case "arm64":
		// ok
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
	// align with install.sh naming scheme
	return fmt.Sprintf("portway-%s-%s", osName, arch), nil
}

func download(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "portway-cli-updater")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func writeTempExecutable(data []byte) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("portway-update-%d", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, data, 0o755); err != nil {
		return "", err
	}
	return tmpFile, nil
}

func replaceCurrentBinary(newPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// On macOS/Linux: write to a sibling file then rename over
	backup := exe + ".bak"
	// remove previous backup if exists
	_ = os.Remove(backup)
	if err := os.Rename(exe, backup); err != nil {
		return fmt.Errorf("failed to backup existing binary: %w", err)
	}
	// move the new binary into place
	if err := os.Rename(newPath, exe); err != nil {
		// try to restore
		_ = os.Rename(backup, exe)
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	// best-effort cleanup
	_ = os.Remove(backup)
	return nil
}

func computeSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func NewUpdateCmd() *cobra.Command {
	var version string
	cmd := &cobra.Command{
		Use:          "update",
		Short:        "Update the CLI to the latest version",
		Long:         "Download and install the latest released version of the Portway CLI.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine version
			target := version
			if strings.TrimSpace(target) == "" {
				v, err := getLatestTag()
				if err != nil {
					return err
				}
				target = v
			}

			name, err := assetName()
			if err != nil {
				return err
			}
			url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", repoOwner, repoName, target, name)

			pterm.Printf("Downloading %s...\n", pterm.Cyan(url))
			data, err := download(url)
			if err != nil {
				return err
			}

			pterm.Printf("SHA256: %s\n", pterm.Gray(computeSHA256(data)))

			tmp, err := writeTempExecutable(data)
			if err != nil {
				return err
			}

			// ensure executable perms
			_ = os.Chmod(tmp, 0o755)

			// If running inside a protected path, we may need sudo. Try direct first.
			if err := replaceCurrentBinary(tmp); err != nil {
				// try sudo move on Unix
				if runtime.GOOS != "windows" {
					exe, exErr := os.Executable()
					if exErr == nil {
						// Move to /usr/local/bin/portway if that's the exe location
						if strings.HasPrefix(exe, "/usr/local/bin/") {
							// sudo mv tmp exe
							mv := exec.Command("sudo", "mv", tmp, exe)
							mv.Stdin = os.Stdin
							mv.Stdout = os.Stdout
							mv.Stderr = os.Stderr
							if mvErr := mv.Run(); mvErr == nil {
								pterm.Success.Println("Updated successfully")
								return nil
							}
						}
					}
				}
				return err
			}

			pterm.Success.Println("Updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "Version tag to install (defaults to latest)")
	return cmd
}
