package lint

import (
	"cli/pkg/api"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/compose-spec/compose-go/v2/types"
)

func Lint(client *api.ClientWithResponses, project *types.Project) ([]api.LintingIssue, error) {
	var result map[string]interface{}
	data, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	response, err := client.LintComposeFileObjectWithResponse(
		context.Background(), result)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("failed to lint compose file")
	}
	return response.JSON200.Results, nil
}

func ConfigLintMessages() error {
	confirmed := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Found linting issues. Do you want to continue?").
				Value(&confirmed),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run prompt: %w", err)
	}

	if !confirmed {
		fmt.Print("\nAborted due to linting issues.\n\n")
		os.Exit(1)
	}

	return nil
}
