package config

import (
	"cli/pkg/api"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

// App represents the main application configuration
type App struct {
	ID   string `yaml:"id"`
	Slug string `yaml:"slug"`
}

// Environment represents configuration for a specific environment
type Environment struct {
	ComposeFiles []string `yaml:"compose_files"`
}

func (e *Environment) GetComposeFiles() ([]string, error) {
	files := []string{}
	
	for _, file := range e.ComposeFiles {
		resolverType := strings.Split(file, ":")[0]

		if len(resolverType) == 1 {
			resolverType = "file"
		}

		resolver, ok := composeFileTypes[resolverType]
		if !ok {
			fmt.Println()
			fmt.Printf("%s compose file type: %s\n", color.RedString("Invalid"), color.YellowString(file))
			fmt.Println()
			return nil, fmt.Errorf("invalid compose file type: %s", file)
		}

		file = strings.TrimPrefix(file, resolverType+":")
		resolvedFile, err := resolver(file)
		if err != nil {
			return nil, err
		}

		files = append(files, resolvedFile)
	}
	return files, nil
}

// Config represents the full application configuration
type Config struct {
	path         string                 `yaml:"-"`
	App          App                    `yaml:"app"`
	Environments map[string]Environment `yaml:"environments"`
}

func (c *Config) GetOrgSlug(client *api.ClientWithResponses) (string, error) {
	whoami, err := client.GetApiV1WhoamiWithResponse(context.Background())
	if err != nil {
		return "", err
	}

	if whoami.StatusCode() != 200 {
		return "", fmt.Errorf("failed to get organization: %d", whoami.StatusCode())
	}

	organization := whoami.JSON200.Organization
	if organization == nil {
		return "", fmt.Errorf("organization not found")
	}

	return organization.Slug, nil
}

func (c *Config) GetAppSlug() string {
	return c.App.Slug
}

// WriteConfig writes the config back to disk at the specified path
func (c *Config) WriteConfig() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func FindConfig(startDir string) (string, error) {
	configNames := []string{".portway.yaml", ".portway.yml"}
	for _, name := range configNames {
		path := filepath.Join(startDir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

// LoadConfig loads and parses the config file at the given path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	config.path = path

	return &config, nil
}

// NewConfig creates a new default config
func NewConfig(path string) *Config {
	return &Config{path: path}
}
