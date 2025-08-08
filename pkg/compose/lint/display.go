package lint

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
)

type ValidationResult struct {
	Errors   []ValidationIssue
	Warnings []ValidationIssue
	Info     []ValidationIssue
}

func DisplayValidationResults(issues []ValidationIssue) {
	totalIssues := len(issues)

	if totalIssues == 0 {
		return
	}

	result := ValidationResult{
		Errors:   []ValidationIssue{},
		Warnings: []ValidationIssue{},
		Info:     []ValidationIssue{},
	}
	for _, issue := range issues {
		switch issue.ValidationCheck.Severity {
		case SeverityError:
			result.Errors = append(result.Errors, issue)
		case SeverityWarning:
			result.Warnings = append(result.Warnings, issue)
		case SeverityInfo:
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

func DisplayIssue(index int, issue ValidationIssue, severityType string) {
	fmt.Printf("   %d. %s %s\n", index, color.CyanString(issue.ValidationCheck.Code), color.New(color.Bold).Sprint(issue.Field))
	fmt.Printf("      %s\n", issue.Message)
	if issue.Suggestion != "" {
		fmt.Printf("      %s %s\n", color.GreenString("💡"), color.New(color.Faint).Sprint(issue.Suggestion))
	}
	pterm.Println()
}
