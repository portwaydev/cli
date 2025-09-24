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

type ComposeFileResolverType string

// Environment represents configuration for a specific environment
type Environment struct {
	Region       string   `yaml:"region,omitempty"`
	Domains      []string `yaml:"domains,omitempty"`
	ComposeFiles []string `yaml:"compose-files"`
}

func (e *Environment) GetComposeFiles(configDir string) ([]string, error) {
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
		resolvedFile, err := resolver(configDir, file)
		if err != nil {
			return nil, err
		}

		files = append(files, resolvedFile)
	}

	return files, nil
}

type ProjectConfig struct {
	DefaultEnvironment string                  `yaml:"default-environment,omitempty"`
	Environments       map[string]*Environment `yaml:"environments"`
}

func (p *ProjectConfig) GetEnvironment(name string) *Environment {
	return p.Environments[name]
}

type GlobalDefaultsConfig struct {
	Region string `yaml:"region,omitempty"`
}

// Config represents the full application configuration
type Config struct {
	path           string                    `yaml:"-"`
	Version        string                    `yaml:"version"`
	DefaultProject string                    `yaml:"default-project"`
	Projects       map[string]*ProjectConfig `yaml:"projects"`

	Defaults *GlobalDefaultsConfig `yaml:"defaults,omitempty"`
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

func (c *Config) GetProjectSlug() string {
	return c.DefaultProject
}

func (c *Config) GetProject() *ProjectConfig {
	if c.DefaultProject != "" {
		if proj, ok := c.Projects[c.DefaultProject]; ok && proj != nil {
			return proj
		}
	}

	for _, proj := range c.Projects {
		if proj != nil {
			return proj
		}
	}

	return nil
}

// WriteConfig writes the config back to disk at the specified path
func (c *Config) WriteConfig() error {
	c.Version = "1.0"

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
	return &Config{path: path, Version: "1.0", Projects: make(map[string]*ProjectConfig)}
}
