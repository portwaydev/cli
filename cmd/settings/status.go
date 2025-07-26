package settings

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Show current settings configuration",
		Long:         "Display the current CLI configuration settings",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pterm.Println(pterm.Green("📋 Current Settings Configuration\n"))

			// Get autoupdate setting
			autoupdate := viper.GetBool("autoupdate")
			var autoupdateStatus string
			if autoupdate {
				autoupdateStatus = pterm.Green("enabled")
			} else {
				autoupdateStatus = pterm.Red("disabled")
			}

			pterm.Printf("Autoupdate: %s\n", autoupdateStatus)

			// Show config file location
			configFile := viper.ConfigFileUsed()
			if configFile != "" {
				pterm.Printf("Config file: %s\n", pterm.Gray(configFile))
			}

			return nil
		},
	}

	return cmd
}
