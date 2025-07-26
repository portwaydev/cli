package auth

import (
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "auth",
		Short:        "Authentication commands",
		Long:         "Commands for managing authentication with Portway",
		SilenceUsage: true,
	}

	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewLogoutCmd())

	return cmd
}
