package validate

import (
	"cli/pkg/compose"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type ValidationResult struct {
	Errors   []ValidationIssue
	Warnings []ValidationIssue
	Info     []ValidationIssue
}

type ValidationIssue struct {
	ValidationCheck *ValidationCheck
	Service         string
	Field           string
	Message         string
	Suggestion      string
}

type ValidationCheck struct {
	Code        string
	Name        string
	Description string
	Severity    string
	Category    string
	CheckFunc   func(ctx ValidationContext) []ValidationIssue
}

type ValidationContext struct {
	Scope   string // global or service
	Project *types.Project
}

func NewValidateCmd() *cobra.Command {
	var composeFile string
	var skipChecks []string

	cmd := &cobra.Command{
		Use:          "validate",
		Short:        "Validate Docker Compose file for Kubernetes deployment",
		Long:         "Analyze a Docker Compose file and identify potential issues when deploying to Kubernetes clusters",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if composeFile == "" {
				// Look for default compose files in current directory
				for _, f := range []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"} {
					if _, err := os.Stat(f); err == nil {
						composeFile = f
						break
					}
				}
			}

			if composeFile == "" {
				pterm.Printf("%s No compose file found\n", pterm.Red("❌"))
				return fmt.Errorf("no compose file found - specify one with -f flag")
			}

			absPath, err := filepath.Abs(composeFile)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			composeConfig, err := compose.LoadComposeConfig([]string{absPath})
			if err != nil {
				return fmt.Errorf("failed to load compose config: %w", err)
			}

			pterm.Printf("Validating compose file: %s\n", pterm.Cyan(absPath))

			// Show skipped checks if any
			if len(skipChecks) > 0 {
				pterm.Printf("Skipping checks: %s\n", pterm.Gray(strings.Join(skipChecks, ", ")))
			}
			pterm.Println()

			// Load and parse the compose file
			result, err := validateComposeFile(composeConfig, skipChecks)
			if err != nil {
				pterm.Printf("%s Failed to validate compose file: %v\n", pterm.Red("❌"), err)
				return err
			}

			// Display results
			displayValidationResults(result)

			// Return error if there are critical errors
			if len(result.Errors) > 0 {
				return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&composeFile, "file", "f", "", "Docker Compose file to validate")
	cmd.Flags().StringSliceVar(&skipChecks, "skip-checks", []string{}, "Skip specific validation checks (comma-separated list of check codes, e.g. PW001,PW002)")

	// Add subcommand to list all checks
	cmd.AddCommand(NewListChecksCmd())

	return cmd
}

func NewListChecksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list-checks",
		Short:        "List all available validation checks",
		Long:         "Display a list of all available validation checks with their codes and descriptions",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			displayAvailableChecks()
			return nil
		},
	}

	return cmd
}

func displayAvailableChecks() {
	pterm.Printf("%s Available Validation Checks\n\n", pterm.Blue("📋"))

	// Create table data
	tableData := pterm.TableData{
		{"Code", "Name", "Description", "Category", "Severity"},
	}

	// Sort checks by code for consistent output
	var codes []string
	for code := range checkRegistry.Checks {
		codes = append(codes, code)
	}

	// Sort the codes
	for i := 0; i < len(codes); i++ {
		for j := i + 1; j < len(codes); j++ {
			if codes[i] > codes[j] {
				codes[i], codes[j] = codes[j], codes[i]
			}
		}
	}

	for _, code := range codes {
		check := checkRegistry.Checks[code]
		var severityColor string
		if check.Severity == "ERROR" {
			severityColor = pterm.Red(check.Severity)
		} else if check.Severity == "INFO" {
			severityColor = pterm.Blue(check.Severity)
		} else {
			severityColor = pterm.Yellow(check.Severity)
		}

		tableData = append(tableData, []string{
			pterm.Cyan(check.Code),
			pterm.Bold.Sprint(check.Name),
			check.Description,
			pterm.Gray(check.Category),
			severityColor,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()

	pterm.Printf("\n%s Usage Examples:\n", pterm.Green("💡"))
	pterm.Printf("  Skip specific checks: %s\n", pterm.Gray("deploy validate --skip-checks PW003,PW005"))
	pterm.Printf("  Skip all warnings: %s\n", pterm.Gray("deploy validate --skip-checks "+getWarningCodes()))
	pterm.Printf("\n%s For detailed documentation: https://docs.portway.dev/validate\n", pterm.Gray("📚"))
}

func getWarningCodes() string {
	var warningCodes []string
	for _, check := range checkRegistry.Checks {
		if check.Severity == "WARNING" {
			warningCodes = append(warningCodes, check.Code)
		}
	}

	// Sort the warning codes
	for i := 0; i < len(warningCodes); i++ {
		for j := i + 1; j < len(warningCodes); j++ {
			if warningCodes[i] > warningCodes[j] {
				warningCodes[i], warningCodes[j] = warningCodes[j], warningCodes[i]
			}
		}
	}

	return strings.Join(warningCodes, ",")
}

func validateComposeFile(composeConfig *types.Project, skipChecks []string) (*ValidationResult, error) {
	result := &ValidationResult{
		Errors:   []ValidationIssue{},
		Warnings: []ValidationIssue{},
		Info:     []ValidationIssue{},
	}

	// Convert skip checks to a map for faster lookup
	skipMap := make(map[string]bool)
	for _, check := range skipChecks {
		skipMap[strings.ToUpper(strings.TrimSpace(check))] = true
	}

	ctx := ValidationContext{Project: composeConfig}
	for _, check := range checkRegistry.Checks {
		shouldSkip := skipMap[check.Code]
		if shouldSkip {
			continue
		}

		issues := check.CheckFunc(ctx)
		for _, issue := range issues {
			issue.ValidationCheck = &check
			if check.Severity == "ERROR" {
				result.Errors = append(result.Errors, issue)
				continue
			}
			if check.Severity == "INFO" {
				result.Info = append(result.Info, issue)
				continue
			}
			if check.Severity == "WARNING" {
				result.Warnings = append(result.Warnings, issue)
				continue
			}
		}
	}

	return result, nil
}

func displayValidationResults(result *ValidationResult) {
	totalIssues := len(result.Errors) + len(result.Warnings) + len(result.Info)

	if totalIssues == 0 {
		pterm.Printf("%s No issues found! Your compose file looks good for Kubernetes deployment.\n", pterm.Green("✅"))
		return
	}

	// Display summary
	pterm.Printf("📊 Validation Summary: %s\n\n",
		pterm.Sprintf("%d issue(s) found", totalIssues))

	// Display errors
	if len(result.Errors) > 0 {
		pterm.Printf("%s Critical Issues (%d):\n", pterm.Red("❌"), len(result.Errors))
		for i, issue := range result.Errors {
			displayIssue(i+1, issue, "ERROR")
		}
		pterm.Println()
	}

	// Display warnings
	if len(result.Warnings) > 0 {
		pterm.Printf("%s  Warnings (%d):\n", pterm.Yellow("⚠️"), len(result.Warnings))
		for i, issue := range result.Warnings {
			displayIssue(i+1, issue, "WARNING")
		}
		pterm.Println()
	}

	// Display info messages
	if len(result.Info) > 0 {
		pterm.Printf("%s Information (%d):\n", pterm.Blue("ℹ️"), len(result.Info))
		for i, issue := range result.Info {
			displayIssue(i+1, issue, "INFO")
		}
		pterm.Println()
	}

	// Summary message
	if len(result.Errors) > 0 {
		pterm.Printf("%s Fix critical issues before deploying to Kubernetes.\n", pterm.Red("🚨"))
	} else if len(result.Warnings) > 0 {
		pterm.Printf("%s Warnings found, but deployment should work. Review suggestions for better reliability.\n", pterm.Yellow("💡"))
	} else {
		pterm.Printf("%s Only informational messages found. Your compose file is ready for deployment.\n", pterm.Green("✅"))
	}

	// Add reference to documentation
	pterm.Printf("\n%s For more information about these checks, visit: https://docs.portway.dev/validate\n", pterm.Gray("📚"))
}

func displayIssue(index int, issue ValidationIssue, severityType string) {
	var severityColor string
	switch severityType {
	case "ERROR":
		severityColor = pterm.Red(severityType)
	case "INFO":
		severityColor = pterm.Blue(severityType)
	default:
		severityColor = pterm.Yellow(severityType)
	}

	pterm.Printf("   %d. [%s] %s %s\n", index, severityColor, pterm.Cyan(issue.ValidationCheck.Code), pterm.Bold.Sprint(issue.Field))
	pterm.Printf("      %s\n", issue.Message)
	if issue.Suggestion != "" {
		pterm.Printf("      %s %s\n", pterm.Green("💡"), pterm.Gray(issue.Suggestion))
	}
	pterm.Println()
}
