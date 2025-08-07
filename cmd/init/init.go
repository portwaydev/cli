package init

import (
	"cli/pkg/api"
	"cli/pkg/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
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

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := api.NewViperClientWithResponses()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			configPath, _ := config.FindConfig(dir)
			if configPath != "" {
				fmt.Printf("\nConfig file already exists at %s\n\n", color.YellowString(configPath))

				overwrite := false
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().Title("Overwrite config file").
							Description("Are you sure you want to overwrite the existing config file?").
							Value(&overwrite),
					),
				)

				form.Run()

				if !overwrite {
					os.Exit(1)
					return nil
				}
			}

			cfg := config.NewConfig(".portway.yaml")

			composeFiles, err := filepath.Glob("*compose*.y*ml")
			if err != nil {
				return err
			}

			composeFileOptions := make([]huh.Option[string], len(composeFiles))
			if len(composeFiles) == 0 {
				fmt.Println("\nNo compose files found in current directory")
				os.Exit(1)
			}

			for i, composeFile := range composeFiles {
				composeFileOptions[i] = huh.NewOption(composeFile, composeFile)
			}

			appSlug := slug.Make(getRepoName())
			composeFile := ""
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Application Identifier").
						Prompt("? ").
						Description("A URL-friendly identifier for your application (e.g. my-cool-app)").
						Value(&appSlug).
						Validate(func(value string) error {
							if len(value) < 2 {
								return fmt.Errorf("application identifier must be at least 2 characters long")
							}
							if !slug.IsSlug(value) {
								return fmt.Errorf("application identifier must be a valid slug (lowercase letters, numbers, and hyphens)")
							}
							return nil
						}),
					huh.NewSelect[string]().
						Title("Compose File").
						Description("The compose file to use for your application").
						Options(composeFileOptions...).
						Value(&composeFile),
				),
			)

			form.Run()

			cfg.App.Slug = appSlug

			environments := make(map[string]config.Environment)

			envName := "production"
			environments[envName] = config.Environment{
				ComposeFiles: []string{"file:" + composeFile},
			}
			cfg.Environments = environments

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

			err = cfg.WriteConfig()
			if err != nil {
				return err
			}

			url := fmt.Sprintf("https://portway.dev/%s/%s", orgSlug, cfg.App.Slug)

			fmt.Printf("View your app at: %s\n", color.BlueString(url))

			return nil
		},
	}

	return cmd
}
