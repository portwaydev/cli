package init

import (
	"cli/pkg/api"
	"cli/pkg/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func getAppSlug() (string, error) {
	repoName := getRepoName()
	repoSlug := slug.Make(repoName)

	appSlug, err := pterm.DefaultInteractiveTextInput.Show("What is the slug for your app? (default: " + repoSlug + ")")
	if err != nil {
		return "", err
	}

	if appSlug == "" {
		appSlug = repoSlug
	}

	return slug.Make(appSlug), nil
}

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			configPath, err := config.FindConfig(dir)
			if configPath != "" {
				return fmt.Errorf("config already exists at %s", configPath)
			}

			cfg := config.NewConfig(".portway.yaml")
			appSlug, err := getAppSlug()
			if err != nil {
				return err
			}

			cfg.App.Slug = slug.Make(appSlug)
			composeFiles, err := filepath.Glob("*compose*.y*ml")
			if err != nil {
				return err
			}

			if len(composeFiles) == 0 {
				fmt.Println("No compose files found in current directory")
				return fmt.Errorf("no compose files found in current directory")
			}

			environments := make(map[string]config.Environment)

			selectedCompose, err := pterm.DefaultInteractiveSelect.WithOptions(composeFiles).Show("Which compose file would you like to configure?")
			if err != nil {
				return err
			}

			envName := "production"
			inputEnvName, err := pterm.DefaultInteractiveTextInput.Show("What is the name of the environment? (default: " + envName + ")")
			if err != nil {
				return err
			}
			if inputEnvName != "" {
				envName = inputEnvName
			}

			environments[envName] = config.Environment{
				ComposeFiles: []string{selectedCompose},
			}

			cfg.Environments = environments

			client, err := api.NewViperClientWithResponses()
			if err != nil {
				return err
			}

			orgSlug, err := cfg.GetOrgSlug(client)
			if err != nil {
				return err
			}

			_, err = client.CreateOrUpdateApp(cmd.Context(), orgSlug, cfg.App.Slug, api.CreateOrUpdateAppJSONRequestBody{
				Name: &cfg.App.Slug,
			})
			if err != nil {
				return err
			}

			url := fmt.Sprintf("https://portway.dev/%s/%s", orgSlug, cfg.App.Slug)
			pterm.Info.Println("View your app at: " + url)

			err = cfg.WriteConfig()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
