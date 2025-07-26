package doctor

import (
	"cli/pkg/api"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "doctor",
		Short:        "Diagnose CLI configuration and connectivity",
		Long:         "Run health checks to diagnose issues with CLI configuration, API connectivity, and authentication",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pterm.Println(pterm.Blue("🔍 Running CLI Health Checks\n"))

			hasErrors := false

			// Check 1: Configuration file
			pterm.Print("📁 Checking configuration file... ")
			configFile := viper.ConfigFileUsed()
			if configFile != "" {
				pterm.Println(pterm.Green("✓"))
				pterm.Printf("   Config file: %s\n", pterm.Gray(configFile))
			} else {
				pterm.Println(pterm.Red("✗"))
				pterm.Printf("   %s No configuration file found\n", pterm.Red("❌"))
				hasErrors = true
			}
			pterm.Println()

			// Check 2: API URL configuration
			pterm.Print("🌐 Checking API URL configuration... ")
			apiURL := viper.GetString("url")
			if apiURL == "" {
				pterm.Println(pterm.Red("✗"))
				pterm.Printf("   %s API URL not configured\n", pterm.Red("❌"))
				hasErrors = true
			} else {
				// Validate URL format
				parsedURL, err := url.Parse(apiURL)
				if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
					pterm.Println(pterm.Red("✗"))
					pterm.Printf("   %s Invalid API URL format: %s\n", pterm.Red("❌"), apiURL)
					hasErrors = true
				} else {
					pterm.Println(pterm.Green("✓"))
					pterm.Printf("   API URL: %s\n", pterm.Gray(apiURL))
				}
			}
			pterm.Println()

			// Check 3: API token/key
			pterm.Print("🔐 Checking API token... ")
			token := viper.GetString("token")
			if token == "" {
				pterm.Println(pterm.Red("✗"))
				pterm.Printf("   %s No API token configured\n", pterm.Red("❌"))
				pterm.Printf("   %s Run 'deploy auth login' to authenticate\n", pterm.Yellow("💡"))
				hasErrors = true
			} else {
				pterm.Println(pterm.Green("✓"))
				// Show token preview (first 8 and last 4 characters)
				var tokenPreview string
				if len(token) > 12 {
					tokenPreview = token[:8] + "..." + token[len(token)-4:]
				} else {
					tokenPreview = strings.Repeat("*", len(token))
				}
				pterm.Printf("   Token: %s\n", pterm.Gray(tokenPreview))
			}
			pterm.Println()

			// Check 4: API connectivity (only if URL and token are available)
			if apiURL != "" && token != "" {
				pterm.Print("🌍 Checking API connectivity... ")

				// Create HTTP client with timeout
				client := &http.Client{
					Timeout: 10 * time.Second,
				}

				// Test basic connectivity
				testURL := strings.TrimSuffix(apiURL, "/") + "/health"
				req, err := http.NewRequest("GET", testURL, nil)
				if err != nil {
					pterm.Println(pterm.Red("✗"))
					pterm.Printf("   %s Failed to create request: %v\n", pterm.Red("❌"), err)
					hasErrors = true
				} else {
					resp, err := client.Do(req)
					if err != nil {
						pterm.Println(pterm.Red("✗"))
						pterm.Printf("   %s Failed to connect: %v\n", pterm.Red("❌"), err)
						pterm.Printf("   %s Check your internet connection and API URL\n", pterm.Yellow("💡"))
						hasErrors = true
					} else {
						resp.Body.Close()
						pterm.Println(pterm.Green("✓"))
						pterm.Printf("   Status: %s\n", pterm.Gray(resp.Status))
					}
				}
				pterm.Println()

				// Check 5: Authentication validity
				pterm.Print("🔑 Checking authentication... ")
				apiClient, err := api.NewAPIKeyClientWithResponses(apiURL, token)
				if err != nil {
					pterm.Println(pterm.Red("✗"))
					pterm.Printf("   %s Failed to create API client: %v\n", pterm.Red("❌"), err)
					hasErrors = true
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					response, err := apiClient.GetApiV1WhoamiWithResponse(ctx)
					if err != nil {
						pterm.Println(pterm.Red("✗"))
						pterm.Printf("   %s Failed to validate authentication: %v\n", pterm.Red("❌"), err)
						hasErrors = true
					} else if response.StatusCode() != 200 {
						pterm.Println(pterm.Red("✗"))
						pterm.Printf("   %s Authentication failed (Status: %d)\n", pterm.Red("❌"), response.StatusCode())
						pterm.Printf("   %s Run 'deploy auth login' to re-authenticate\n", pterm.Yellow("💡"))
						hasErrors = true
					} else {
						pterm.Println(pterm.Green("✓"))
						if response.JSON200 != nil && response.JSON200.Organization != nil {
							org := response.JSON200.Organization
							pterm.Printf("   Organization: %s (%s)\n", pterm.Gray(org.Name), pterm.Gray(org.Slug))
						}
					}
				}
				pterm.Println()
			}

			// Check 6: Settings configuration
			pterm.Print("⚙️  Checking settings... ")
			autoupdate := viper.GetBool("autoupdate")
			pterm.Println(pterm.Green("✓"))
			var autoupdateStatus string
			if autoupdate {
				autoupdateStatus = pterm.Green("enabled")
			} else {
				autoupdateStatus = pterm.Gray("disabled")
			}
			pterm.Printf("   Autoupdate: %s\n", autoupdateStatus)
			pterm.Println()

			// Summary
			if hasErrors {
				pterm.Printf("%s Some issues were found. Please address them before using the CLI.\n", pterm.Red("❌"))
				return fmt.Errorf("health check failed")
			} else {
				pterm.Printf("%s All checks passed! Your CLI is ready to use.\n", pterm.Green("✅"))
			}

			return nil
		},
	}

	return cmd
}
