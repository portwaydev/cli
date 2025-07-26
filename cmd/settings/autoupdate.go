package settings

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewAutoupdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "autoupdate <enable|disable>",
		Short:        "Enable or disable automatic updates",
		Long:         "Enable or disable automatic updates for the CLI. When enabled, the CLI will check for updates and prompt to install them.",
		Args:         cobra.ExactArgs(1),
		ValidArgs:    []string{"enable", "disable"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			action := strings.ToLower(args[0])

			switch action {
			case "enable":
				viper.Set("autoupdate", true)
				if err := viper.WriteConfig(); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}
				pterm.Success.Println("Autoupdate enabled")
				pterm.Info.Println("The CLI will check for updates and prompt to install them")

			case "disable":
				viper.Set("autoupdate", false)
				if err := viper.WriteConfig(); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}
				pterm.Success.Println("Autoupdate disabled")
				pterm.Info.Println("The CLI will not check for updates automatically")

			default:
				return fmt.Errorf("invalid argument: %s. Use 'enable' or 'disable'", action)
			}

			return nil
		},
	}

	return cmd
}
