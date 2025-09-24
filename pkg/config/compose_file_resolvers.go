package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// composeFileTypes maps compose file type prefixes to their resolver functions
var composeFileTypes = map[string]ComposeFileResolver{
	"github": GetGithubComposeFile,
	"file":   GetFileComposeFile,
	"url":    GetHTTPComposeFile,
}

// ComposeFileResolver is a function that resolves a compose file reference to a local file path
type ComposeFileResolver func(configDir string, ref string) (string, error)

// GetFileComposeFile resolves a local file reference by returning the path as-is
func GetFileComposeFile(configDir string, ref string) (string, error) {
	fp, err := filepath.Abs(filepath.Join(configDir, ref))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		fmt.Println()
		fmt.Printf("%s compose file not found at %s\n", color.RedString("Error:"), color.YellowString(ref))
		fmt.Println()
		return "", fmt.Errorf("compose file not found: %s", ref)
	}

	return fp, nil
}

// GetGithubComposeFile downloads a compose file from GitHub and returns the local temp file path
// Expected format: owner/repo/path/to/file@version
func GetGithubComposeFile(_ string, ref string) (string, error) {
	// Parse reference into repo path and version
	refParts := strings.Split(ref, "@")
	if len(refParts) != 2 {
		return "", fmt.Errorf("invalid github compose file format: %s", ref)
	}

	fullRepoPath := refParts[0]
	gitRef := refParts[1]

	// Extract repo and file path components
	pathComponents := strings.Split(fullRepoPath, "/")
	if len(pathComponents) < 3 {
		return "", fmt.Errorf("invalid github repo path: %s", fullRepoPath)
	}

	repoOwnerAndName := strings.Join(pathComponents[:2], "/")
	filePath := strings.Join(pathComponents[2:], "/")

	// Create temporary file for downloaded content
	tempFile, err := os.CreateTemp("", "github-compose-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Build GitHub raw content URL
	downloadURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repoOwnerAndName, gitRef, filePath)

	// Create HTTP client with reasonable timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download file from GitHub
	response, err := httpClient.Get(downloadURL)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer response.Body.Close()

	// Handle non-200 status codes with detailed error output
	if response.StatusCode != http.StatusOK {
		os.Remove(tempFile.Name())
		fmt.Println()
		fmt.Printf(
			"%s to download GitHub file.\n",
			color.RedString("Failed"),
		)
		fmt.Printf("Reference: %s\n", color.CyanString(ref))
		fmt.Printf("Status: %s\n", color.YellowString(strconv.Itoa(response.StatusCode)))
		fmt.Printf("URL: %s\n", color.YellowString(downloadURL))
		fmt.Println()
		return "", fmt.Errorf("failed to download file, status: %d", response.StatusCode)
	}

	// Write downloaded content to temporary file
	_, err = io.Copy(tempFile, response.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return tempFile.Name(), nil
}

func GetHTTPComposeFile(_ string, ref string) (string, error) {
	response, err := http.Get(ref)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	tempFile, err := os.CreateTemp("", "http-compose-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tempFile, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return tempFile.Name(), nil
}
