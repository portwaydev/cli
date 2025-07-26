package launch

import (
	"fmt"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/spf13/cobra"
)

func NewLaunchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "launch",
		Short: "Launch a new service",
		RunE: func(cmd *cobra.Command, args []string) error {
			composeFilePath := "docker-compose.yml"
			options, err := cli.NewProjectOptions(
				[]string{composeFilePath},
			)
			if err != nil {
				return err
			}

			project, err := options.LoadProject(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Println(project.Services)

			return nil
		},
	}
}
