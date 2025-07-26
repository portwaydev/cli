package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// PullImage pulls the specified Docker image
func PullImage(image string) error {
	cmd := exec.Command("docker", "pull", image)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	return nil
}

// TagImage tags a Docker image with a new name
func TagImage(sourceImage string, targetImage string) error {
	cmd := exec.Command("docker", "tag", sourceImage, targetImage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image %s: %w", sourceImage, err)
	}
	return nil
}

// PushImage pushes a Docker image to a registry
func PushImage(image string) error {
	fmt.Printf("Pushing image %s\n", image)
	cmd := exec.Command("docker", "push", image)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push image %s: %w", image, err)
	}
	return nil
}

// GetImageID returns the SHA identifier for a Docker image
func GetImageID(image string) (string, error) {
	inspectCmd := exec.Command("docker", "inspect", "--format", "{{.Id}}", image)
	output, err := inspectCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get image ID for %s: %w", image, err)
	}

	imageID := strings.TrimSpace(string(output))
	if imageID == "" {
		return "", fmt.Errorf("could not determine ID for image %s", image)
	}

	return imageID[7:16], nil
}
