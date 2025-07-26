package compose

import (
	"context"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

func LoadComposeConfig(configs []string) (*types.Project, error) {
	opts, err := cli.NewProjectOptions(
		configs,
		cli.WithOsEnv,
		cli.WithDotEnv,
	)
	if err != nil {
		return nil, err
	}

	return opts.LoadProject(context.Background())
}
