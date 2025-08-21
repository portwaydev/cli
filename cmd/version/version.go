package version

import (
	"fmt"

	"cli/pkg/buildinfo"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:          "version",
        Short:        "Show CLI version information",
        SilenceUsage: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(pterm.TableData{
                {"Field", "Value"},
                {"Version", buildinfo.Version},
                {"Commit", buildinfo.Commit},
                {"Date", buildinfo.Date},
            }).Render()
            fmt.Println()
            return nil
        },
    }
    return cmd
}


