package lint

import (
	"cli/pkg/api"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
)

type ValidationResult struct {
	Errors   []api.LintingIssue
	Warnings []api.LintingIssue
	Info     []api.LintingIssue
}

func DisplayValidationResults(issues []api.LintingIssue) {
	totalIssues := len(issues)

	if totalIssues == 0 {
		return
	}

	result := ValidationResult{
		Errors:   []api.LintingIssue{},
		Warnings: []api.LintingIssue{},
		Info:     []api.LintingIssue{},
	}

	for _, issue := range issues {
		switch strings.ToLower(string(issue.Severity)) {
		case "error":
			result.Errors = append(result.Errors, issue)
		case "warning":
			result.Warnings = append(result.Warnings, issue)
		case "info":
			result.Info = append(result.Info, issue)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Printf("%s  Critical Issues (%d):\n", color.RedString("❌"), len(result.Errors))
		fmt.Println()
		for i, issue := range result.Errors {
			DisplayIssue(i+1, issue, "ERROR")
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Printf("%s  Warning Issues (%d):\n", color.YellowString("⚠️"), len(result.Warnings))
		fmt.Println()
		for i, issue := range result.Warnings {
			DisplayIssue(i+1, issue, "WARNING")
		}
		fmt.Println()
	}

	if len(result.Info) > 0 {
		fmt.Println()
		fmt.Printf("%s  Info Issues (%d):\n", color.BlackString("ℹ️"), len(result.Info))
		fmt.Println()
		for i, issue := range result.Info {
			DisplayIssue(i+1, issue, "INFO")
		}
		pterm.Println()
	}
}

func DisplayIssue(index int, issue api.LintingIssue, severityType string) {
	fmt.Printf("   %d. %s %s\n", index, color.CyanString(issue.Code), color.New(color.Bold).Sprint(issue.Scope))
	fmt.Printf("      %s\n", issue.Message)
	if issue.Context != nil {
		fmt.Printf("      %s\n", color.New(color.Faint).Sprint(*issue.Context))
	}
	if issue.DocUrl != "" {
		fmt.Printf("      %s %s\n", color.GreenString("💡"), color.New(color.Faint).Sprint(issue.DocUrl))
	}
	pterm.Println()
}
