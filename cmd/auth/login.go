package auth

import (
	"cli/pkg/util"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewLoginCmd() *cobra.Command {
	var apiKey string
	var relogin bool
	var token string

	cmd := &cobra.Command{
		Use:          "login",
		Short:        "Login to the Portway CLI",
		Long:         "Authenticate with Portway by opening a browser and logging in",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token != "" {
				viper.Set("token", strings.TrimSpace(token))
				if err := viper.WriteConfig(); err != nil {
					return fmt.Errorf("failed to save token: %w", err)
				}
				fmt.Println("✅ Authentication token set successfully!")
				return nil
			}

			if util.IsCI() {
				return fmt.Errorf("Authentication requires --token flag in CI environments (browser-based auth not available)")
			}

			// Check if already logged in
			if viper.GetString("token") != "" && !relogin {
				fmt.Println("Already logged in. Use --relogin flag to force a new login")
				return nil
			}

			// Get random port between 49152-65535 (ephemeral port range)
			port := 49152 + rand.Intn(65535-49152+1)
			token, err := startLocalServer(cmd.Context(), port)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			pterm.Println(pterm.Green("✅ Authentication successful!"))
			viper.Set("token", token)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "Token to use for authentication")
	cmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key to use for authentication")
	cmd.Flags().BoolVar(&relogin, "relogin", false, "Force a new login even if already logged in")

	return cmd
}
