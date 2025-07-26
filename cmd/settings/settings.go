package settings

import (
	"github.com/spf13/cobra"
)

func NewSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "settings",
		Short:        "Configuration settings commands",
		Long:         "Commands for managing CLI configuration settings",
		SilenceUsage: true,
	}

	cmd.AddCommand(NewAutoupdateCmd())
	cmd.AddCommand(NewStatusCmd())

	return cmd
}
