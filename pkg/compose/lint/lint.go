package lint

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/compose-spec/compose-go/v2/types"
)

type Severity string
const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

type ValidationCheck struct {
	Code        string
	Name        string
	Description string
	Severity    Severity
	Category    string
	CheckFunc   func(ctx *types.Project) []ValidationIssue
}

type ValidationIssue struct {
	ValidationCheck *ValidationCheck
	Service         string
	Field           string
	Message         string
	Suggestion      string
}

var checks = []ValidationCheck{
	ServiceKeyRFC1123,
}

func Lint(project *types.Project) []ValidationIssue {
	var issues []ValidationIssue
	for _, check := range checks {
		for _, issue := range check.CheckFunc(project) {
			issue.ValidationCheck = &check
			issues = append(issues, issue)
		}
	}
	return issues
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
