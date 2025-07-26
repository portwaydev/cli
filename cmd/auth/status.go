package auth

import (
	"cli/pkg/api"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Show authentication status",
		Long:         "Display the current authentication status and token information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			token := viper.GetString("token")

			if token == "" {
				pterm.Error.Println("Not authenticated")
				pterm.Info.Println("Run 'deploy auth login' to authenticate with browser")
				pterm.Info.Println("Or run 'deploy auth set-token --token <your-token>' to set token manually")
				return nil
			}

			client, err := api.NewAPIKeyClientWithResponses(viper.GetString("url"), token)
			if err != nil {
				return err
			}

			response, err := client.GetApiV1WhoamiWithResponse(cmd.Context())
			if err != nil {
				return err
			}

			if response.StatusCode() != 200 {
				pterm.Error.Println("Not authenticated")
				pterm.Info.Println("Run 'deploy auth login' to authenticate with browser")
				pterm.Info.Println("Or run 'deploy auth set-token --token <your-token>' to set token manually")
				return nil
			}
			organization := response.JSON200.Organization

			pterm.Println(pterm.Green("✅ Authenticated\n"))
			pterm.Printf("Organization: %s (%s)\n", pterm.Cyan(organization.Name), pterm.Gray(organization.Slug))
			pterm.Println()

			// Show token preview (first 8 and last 4 characters)
			var tokenPreview string
			if len(token) > 12 {
				tokenPreview = token[:8] + "..." + token[len(token)-4:]
			} else {
				tokenPreview = strings.Repeat("*", len(token))
			}

			pterm.Printf("Token: %s\n", pterm.Cyan(tokenPreview))

			configFile := viper.ConfigFileUsed()
			if configFile != "" {
				pterm.Printf("Config: %s\n", pterm.Gray(configFile))
			}

			return nil
		},
	}

	return cmd
}
