package auth

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "logout",
		Short:        "Clear authentication token",
		Long:         "Remove the stored authentication token and tenant information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			token := viper.GetString("token")

			if token == "" {
				pterm.Info.Println("Already logged out")
				return nil
			}

			// Clear authentication data
			viper.Set("token", "")
			viper.Set("tenant", "")

			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			pterm.Success.Println("Successfully logged out")
			pterm.Info.Println("Authentication token has been cleared")

			return nil
		},
	}

	return cmd
}
