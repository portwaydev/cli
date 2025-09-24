package deploy

import (
	"bytes"
	initcmd "cli/cmd/init"
	"cli/pkg/compose"
	"cli/pkg/compose/lint"
	"cli/pkg/config"
	"cli/pkg/docker"
	"cli/pkg/util"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cli/pkg/api"

	"github.com/charmbracelet/huh"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printHealth(client *api.ClientWithResponses, deploymentId uuid.UUID) error {
	health, err := client.GetDeploymentHealthWithResponse(
		context.Background(),
		deploymentId,
	)
	if err != nil {
		return fmt.Errorf("failed to get deployment health: %w", err)
	}

	if health.StatusCode() != 200 {
		return fmt.Errorf("failed to get deployment health")
	}

	info := health.JSON200

	fmt.Println(info.Summary)
	fmt.Println()

	fmt.Println()
	podTableData := pterm.TableData{{"Pod", "Phase", "Restarts"}}
	for _, pod := range *info.Health.Pods {
		podTableData = append(podTableData, []string{
			*pod.Name,
			*pod.Phase,
			fmt.Sprintf("%d", int(*pod.Restarts)),
		})
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(podTableData).Render()
	
	resourcesTableData := pterm.TableData{{"Service", "Ready", "Desired"}}
	for _, service := range *info.Health.Resources.Deployments {
		resourcesTableData = append(resourcesTableData, []string{
			*service.Name, 
			fmt.Sprintf("%d", int(*service.Ready)),
			fmt.Sprintf("%d", int(*service.Desired)),
		})
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(resourcesTableData).Render()

	if len(*info.Troubleshooting.Suggestions) > 0 {
		fmt.Println()
		fmt.Println(color.YellowString("Suggestions:"))
		for _, suggestion := range *info.Troubleshooting.Suggestions {
			fmt.Println(color.YellowString(suggestion))
		}
	}

	return nil
}

func createEnvironmentComposeFile(
	client *api.ClientWithResponses,
	cfg *config.Config,
	envName string,
	version string,
	composeConfig *types.Project,
) (*api.CreateEnvironmentComposeFileResponse, error) {
	jsonCompose, err := composeConfig.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compose config: %w", err)
	}

	yamlCompose, err := composeConfig.MarshalYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compose config: %w", err)
	}

	composeConfigMap := make(map[string]any)
	err = json.Unmarshal(jsonCompose, &composeConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal compose config: %w", err)
	}

	orgSlug, err := cfg.GetOrgSlug(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization slug: %w", err)
	}

	raw := string(yamlCompose)
	return client.CreateEnvironmentComposeFileWithResponse(
		context.Background(),
		orgSlug,
		cfg.GetProjectSlug(),
		envName,
		api.CreateEnvironmentComposeFileJSONRequestBody{
			ComposeNoramlized: &composeConfigMap,
			ComposeRaw:        &raw,
			Version:           version,
		},
	)
}

func getConfig(configPath string, cmd *cobra.Command, args []string) (*config.Config, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		err = initcmd.NewInitCmd().RunE(cmd, args)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", err)
		}
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return cfg, nil
}

func printServicesTable(composeConfig *types.Project) {
	fmt.Println()
	pterm.Printf("Found %s services\n\n", pterm.Cyan(fmt.Sprintf("%d", len(composeConfig.Services))))

	tableData := pterm.TableData{{"Service", "Image", "Build Context"}}
	// Get all service names and sort them
	serviceNames := make([]string, 0, len(composeConfig.Services))
	for serviceName := range composeConfig.Services {
		serviceNames = append(serviceNames, serviceName)
	}
	sort.Strings(serviceNames)

	// Add services to table in sorted order
	for _, serviceName := range serviceNames {
		service := composeConfig.Services[serviceName]
		buildStatus := pterm.Red("No")
		if service.Build != nil {
			buildStatus = pterm.Green("Yes")
		}
		tableData = append(tableData, []string{
			pterm.Bold.Sprint(serviceName),
			pterm.Cyan(service.Image),
			buildStatus,
		})
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
}

func NewDeployCmd() *cobra.Command {
	var configPath string
	var envName string
	var version string
	var force bool

	cmd := &cobra.Command{
		Use:          "deploy",
		Short:        "Deploy a docker compose file",
		Long:         "Deploy a docker compose",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := filepath.Dir(configPath)
			if hasUncommittedChanges(configDir) {
				if !util.IsCI() {
					if err := confirmUncommittedChanges(); err != nil {
						return err
					}
				} else {
					pterm.Printf("%s  Warning: There are uncommitted changes\n", pterm.Yellow("⚠️"))
				}
			}

			client, err := api.NewViperClientWithResponses()
			if err != nil {
				fmt.Println()
				fmt.Printf("Failed to create client.\n")
				fmt.Printf("Error message: %s\n", color.RedString(err.Error()))
				fmt.Println()
				fmt.Printf("Please check your API key and try again.\n\n")
				os.Exit(1)
			}

			cfg, err := getConfig(configPath, cmd, args)
			if err != nil {
				return err
			}

			orgSlug, err := cfg.GetOrgSlug(client)
			if err != nil {
				return err
			}

			name := cfg.GetProjectSlug()
			app, err := client.CreateOrUpdateAppWithResponse(cmd.Context(), orgSlug, cfg.GetProjectSlug(), api.CreateOrUpdateAppJSONRequestBody{
				Name: &name,
			})
			if err != nil {
				return err
			}

			appID := app.JSON200.Id.String()

			project := cfg.GetProject()
			if project == nil {
				return fmt.Errorf("environment %s not found in project", envName)
			}

			env := project.GetEnvironment(envName)
			if envName == "" || env == nil {
				return fmt.Errorf("no environment specified or found in config")
			}

			composeFiles, err := env.GetComposeFiles(configDir)
			if err != nil {
				pterm.Printf("%s Failed to get compose files\n", pterm.Red("❌"))
				return fmt.Errorf("failed to get compose files: %w", err)
			}

			if len(composeFiles) == 0 {
				fmt.Println()
				fmt.Println(color.RedString("No compose files found."))
				fmt.Println()
				os.Exit(1)
			}

			composeConfig, err := compose.LoadComposeConfig(composeFiles)
			if err != nil {
				pterm.Printf("%s Failed to load compose config\n", pterm.Red("❌"))
				return fmt.Errorf("failed to get services with build: %w", err)
			}

			issues, err := lint.Lint(client, composeConfig)
			if err != nil {
				return fmt.Errorf("failed to lint compose config: %w", err)
			}

			if len(issues) > 0 {
				lint.DisplayValidationResults(issues)
			}

			if !force {
				for _, issue := range issues {
					if strings.ToLower(string(issue.Severity)) == "error" {
						os.Exit(1)
					}
				}
			}

			if len(issues) > 0 {
				lint.ConfigLintMessages()
			}

			printServicesTable(composeConfig)

			// Building images
			hasImagesToBuild := len(composeConfig.ServicesWithBuild()) > 0
			if hasImagesToBuild {
				fmt.Println("🔨 Building all images defined in the compose file...")
				cmdArgs := []string{"compose"}
				for _, file := range composeFiles {
					cmdArgs = append(cmdArgs, "-f", file)
				}
				cmdArgs = append(cmdArgs, "build")

				var stdoutBuf, stderrBuf bytes.Buffer
				
				cmd := exec.Command("docker", cmdArgs...)
				cmd.Stdout = &stdoutBuf
				cmd.Stderr = &stderrBuf
				cmd.Env = append(os.Environ(), "DOCKER_DEFAULT_PLATFORM=linux/amd64")

				buildSpinner := NewSpinnerWithText("Building")
				go func() {
					_, _ = buildSpinner.Run()
				}()

				err := cmd.Run()
				if err != nil {
					ExitSpinner(buildSpinner, color.RedString("Failed to build images."))
					// Print captured logs for debugging
					if stdoutBuf.Len() > 0 {
						fmt.Println(color.YellowString("--- docker build stdout ---"))
						fmt.Print(stdoutBuf.String())
					}
					if stderrBuf.Len() > 0 {
						fmt.Println(color.YellowString("--- docker build stderr ---"))
						fmt.Print(stderrBuf.String())
					}
					fmt.Println("docker", strings.Join(cmdArgs, " "))
					fmt.Printf("%s Failed to build images: %v\n", pterm.Red("❌"), err)
					return fmt.Errorf("failed to build images: %w", err)
				}

				ExitSpinner(buildSpinner, "Docker images built.")

				fmt.Println()
				pterm.Println("ℹ️  Built Image IDs")
				serviceImages := map[string]string{}
				for serviceName, service := range composeConfig.Services {
					if service.Build == nil {
						continue
					}

					candidates := []string{}
					if service.Image != "" {
						candidates = append(candidates, service.Image, service.Image+":latest")
					} else {
						baseUnderscore := fmt.Sprintf("%s_%s", composeConfig.Name, serviceName)
						baseHyphen := fmt.Sprintf("%s-%s", composeConfig.Name, serviceName)
						candidates = append(candidates,
							baseUnderscore,
							baseUnderscore+":latest",
							baseHyphen,
							baseHyphen+":latest",
						)
					}

					var foundRef string
					var imageID string
					for _, ref := range candidates {
						id, idErr := docker.GetImageID(ref)
						if idErr == nil {
							foundRef = ref
							imageID = id
							break
						}
					}

					if foundRef == "" {
						pterm.Printf("%s Could not determine image ID for %s\n", pterm.Yellow("⚠️"), pterm.Bold.Sprint(serviceName))
						continue
					}

					pterm.Printf("🏷️  %s → %s (%s)\n", pterm.Bold.Sprint(serviceName), pterm.Cyan(foundRef), pterm.Green(imageID))

					newRef := fmt.Sprintf("registry.portway.dev/%s/%s:%s-%s", appID, serviceName, envName, imageID)

					if err := docker.TagImage(foundRef, newRef); err != nil {
						pterm.Printf("%s Failed to retag: %s → %s (%s)\n", pterm.Red("❌"), pterm.Cyan(foundRef), pterm.Green(newRef), err.Error())
						os.Exit(1)
					}
					pterm.Printf("  Retagged to %s\n", pterm.Cyan(newRef))
					serviceImages[serviceName] = newRef

					serviceCopy := service
					serviceCopy.Image = newRef
					composeConfig.Services[serviceName] = serviceCopy
				}
				fmt.Println()

				// Push images to the distributed registry, passing in the API key
				pterm.Println("🚀 Pushing images to registry...")
				apiKey := viper.GetString("token")
				if strings.TrimSpace(apiKey) == "" {
					pterm.Printf("%s Missing API token. Please set it with 'portway auth login' or configure 'token' in config.\n", pterm.Red("❌"))
					os.Exit(1)
				}
				loginCmd := exec.Command("docker", "login", "registry.portway.dev", "-u", "portway", "--password-stdin")
				loginCmd.Stdin = strings.NewReader(apiKey)
				loginCmd.Stdout = os.Stdout
				loginCmd.Stderr = os.Stderr
				if err := loginCmd.Run(); err != nil {
					pterm.Printf("%s Failed to login to registry. Please verify your API token.\n", pterm.Red("❌"))
					os.Exit(1)
				}

				fmt.Println()

				for _, imageRef := range serviceImages {
					pterm.Printf("📤 Pushing %s...\n", pterm.Cyan(imageRef))
					if err := docker.PushImage(imageRef); err != nil {
						pterm.Printf("%s Failed to push image %s: %s\n", pterm.Red("❌"), pterm.Cyan(imageRef), err.Error())
						os.Exit(1)
					}
					pterm.Printf("%s Successfully pushed %s\n", pterm.Green("✅"), pterm.Cyan(imageRef))
				}

				fmt.Println()
			}

			if version == "" {
				version, _ = determineVersion()

				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Enter a version").
							Value(&version).
							Placeholder("Enter a version"),
					),
				)

				err := form.Run()
				if err != nil {
					return fmt.Errorf("failed to get version input: %w", err)
				}
			}

			composeFileResponse, err := createEnvironmentComposeFile(
				client,
				cfg,
				envName,
				version,
				composeConfig,
			)
			if err != nil {
				return fmt.Errorf("failed to create compose file: %w", err)
			}

			if composeFileResponse.StatusCode() != 200 {
				fmt.Println()
				color.Red("Failed to create compose file.")
				fmt.Printf("Status: %s\n", color.YellowString(strconv.Itoa(composeFileResponse.StatusCode())))
				fmt.Println()
				return fmt.Errorf("failed to create compose file")
			}

			fmt.Println()
			fmt.Printf("Created version %s of compose file.\n", color.GreenString(version))
			fmt.Println()

			deployResponse, err := client.DeployEnvironmentComposeFileWithResponse(
				context.Background(),
				composeFileResponse.JSON200.Id,
			)

			if err != nil {
				return fmt.Errorf("failed to deploy environment compose file: %w", err)
			}

			if deployResponse.StatusCode() != 200 {
				fmt.Println()
				color.Red("Failed to deploy environment compose file.\n")
				fmt.Println()
				return fmt.Errorf("failed to deploy environment compose file")
			}

			deployments := deployResponse.JSON200.Deployments

			if len(deployments) == 0 {
				fmt.Println()
				color.Yellow("No deployment targets found.")
				fmt.Println("This can happen if you have deleted existing targets, have no branches configured, or have not set up any deployment targets.")
				fmt.Println()
				return nil
			}

			spinner := NewSpinner()
			
			// Channel to signal spinner completion/interruption
			done := make(chan error, 1)
			
			go func() {
				// Run spinner in background and capture if it was interrupted
				model, err := spinner.Run()
				if err != nil {
					done <- err
					return
				}
				
				// Check if the spinner was quitting (possibly due to Ctrl+C)
				if spinnerModel, ok := model.(spinnerModel); ok && spinnerModel.quitting {
					done <- fmt.Errorf("operation interrupted by user")
					return
				}
				
				done <- nil
			}()

			// Wait for all deployments to complete or spinner interruption
			for _, d := range deployments {
				for {
					// Check if spinner was interrupted
					select {
					case err := <-done:
						if err != nil {
							fmt.Println()
							fmt.Println(color.RedString("Deployment interrupted."))
							fmt.Println()
							return err
						}
					default:
						// Continue with deployment check
					}

					deployment, err := client.GetDeploymentWithResponse(
						context.Background(),
						d.Id.String(),
					)

					if err != nil {
						ExitSpinner(spinner, color.RedString("Deployment failed."))
						fmt.Println()
						return fmt.Errorf("failed to get deployment: %w", err)
					}

					logs := *deployment.JSON200.Logs
					for _, l := range logs {
						log := logMsg{
							timestamp: l.Timestamp,
							log:       l.Log,
							stream:    l.Stream,
						}
						spinner.Send(log)
					}

					if deployment.JSON200.Status == "failed" {
						message := "An error happened while trying to deploy your application."
						ExitSpinner(spinner, color.RedString(message))
						fmt.Println()
						err = printHealth(client, *d.Id)
						if err != nil {
							fmt.Println()
							fmt.Println(color.RedString("Failed to print health."))
							fmt.Println(color.RedString(err.Error()))
							fmt.Println()
						}
						return fmt.Errorf("deployment failed")
					}

					if deployment.JSON200.Status == "deployed" {
						break
					}

					// Sleep briefly before checking again
					time.Sleep(1 * time.Second)
				}
			}

			ExitSpinner(spinner, "Deployments completed.")
			fmt.Println()
			fmt.Println()

			printHealth(client, *deployments[0].Id)

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", ".portway.yaml", "Config file to deploy")
	cmd.Flags().StringVarP(&envName, "env", "e", "", "Environment to deploy to")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version to deploy")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deploy")

	return cmd
}
