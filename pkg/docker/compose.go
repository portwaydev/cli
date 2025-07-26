package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
)

func LoadComposeConfig(composeFile string) (*ComposeConfig, error) {
	configBytes, err := getComposeConfig(composeFile)
	if err != nil {
		return nil, err
	}
	config, err := parseComposeConfig(configBytes)
	if err != nil {
		return nil, err
	}
	config.composeFile = composeFile
	return config, nil
}

// Service represents a service in the Docker Compose configuration
type Service struct {
	Image string         `json:"image"`
	Build map[string]any `json:"build"`
}

// ComposeConfig represents the full Docker Compose configuration
type ComposeConfig struct {
	composeFile string             `json:"-"`
	Project     string             `json:"project"`
	Services    map[string]Service `json:"services"`
}

func (c *ComposeConfig) GetServicesWithBuild() []string {
	var services []string
	for serviceName, service := range c.Services {
		if service.Build != nil {
			services = append(services, serviceName)
		}
	}
	return services
}

func (c *ComposeConfig) GetLocalImages() []string {
	var images []string
	for serviceName, service := range c.Services {
		// Only include images that have a build section (locally built)
		if service.Build != nil {
			if service.Image != "" {
				images = append(images, service.Image)
			} else {
				// If no image name specified, use service name with project prefix
				configName := "default"
				if c.Project != "" {
					configName = c.Project
				}
				images = append(images, fmt.Sprintf("%s_%s", configName, serviceName))
			}
		}
	}
	return images
}

func (c *ComposeConfig) GetExternalImages() []string {
	var images []string
	for _, service := range c.Services {
		if service.Build == nil && service.Image != "" {
			images = append(images, service.Image)
		}
	}
	return images
}

func (c *ComposeConfig) GetImages() []string {
	var images []string
	for _, service := range c.Services {
		if service.Image != "" {
			images = append(images, service.Image)
			continue
		}

		configName := "default"
		if c.Project != "" {
			configName = c.Project
		}

		images = append(images, fmt.Sprintf("%s_%s", configName, service.Image))
	}

	return images
}

func (c *ComposeConfig) Build() error {
	buildCmd := exec.Command("docker", "compose", "-p", c.Project, "-f", c.composeFile, "build")
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build images: %w", err)
	}
	return nil
}

// ParseComposeConfig parses the JSON compose config into a struct
func parseComposeConfig(configBytes []byte) (*ComposeConfig, error) {
	var config ComposeConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse compose config: %w", err)
	}
	return &config, nil
}

// GetComposeConfig gets the Docker Compose configuration in JSON format
func getComposeConfig(composeFile string) ([]byte, error) {
	composeCmd := exec.Command("docker", "compose", "-f", composeFile, "config", "--format", "json")
	output, err := composeCmd.CombinedOutput()
	if err != nil {
		pterm.Error.Println(string(output))
		return output, fmt.Errorf("failed to get compose config: %w", err)
	}
	return output, nil
}
