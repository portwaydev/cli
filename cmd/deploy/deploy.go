package deploy

import (
	initcmd "cli/cmd/init"
	"cli/pkg/compose"
	"cli/pkg/config"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"cli/pkg/api"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// func createComposeFile(client *api.ClientWithResponses, composeFileName string, organizationId uuid.UUID) (uuid.UUID, error) {
// 	composeFileRequestBody := api.UpsertComposeFileV1JSONRequestBody{
// 		Name: &composeFile,
// 	}
// 	composeFileResponse, err := client.UpsertComposeFileV1WithResponse(
// 		context.Background(),
// 		organizationId,
// 		composeFile,
// 		composeFileRequestBody,
// 	)

// 	if err != nil {
// 		pterm.Printf("%s Failed to upsert compose file\n", pterm.Red("❌"))
// 		return uuid.Nil, fmt.Errorf("failed to upsert compose file: %w", err)
// 	}

// 	if composeFileResponse.JSON200 != nil {
// 		return composeFileResponse.JSON200.Id, nil
// 	}
// 	if composeFileResponse.JSON201 != nil {
// 		return composeFileResponse.JSON201.Id, nil
// 	}

// 	pterm.Printf("%s Failed to get compose file ID\n", pterm.Red("❌"))
// 	return uuid.Nil, fmt.Errorf("failed to get compose file ID")
// }

// func createComposeFileVersion(client *api.ClientWithResponses, composeFileId uuid.UUID, composeFile string) (uuid.UUID, error) {
// 	yamlBytes, err := os.ReadFile(composeFile)
// 	if err != nil {
// 		pterm.Printf("%s Failed to read compose file\n", pterm.Red("❌"))
// 		return uuid.Nil, fmt.Errorf("failed to read compose file: %w", err)
// 	}

// 	version, err := determineVersion(composeFile)
// 	if err != nil {
// 		return uuid.Nil, fmt.Errorf("failed to determine version: %w", err)
// 	}
// 	source := api.CreateComposeFileVersionV1JSONBodySourceCli
// 	composeFileVersionRequestBody := api.CreateComposeFileVersionV1JSONRequestBody{
// 		RawCompose: string(yamlBytes),
// 		Version:    version,
// 		Source:     &source,
// 	}

// 	composeFileVersionResponse, err := client.CreateComposeFileVersionV1WithResponse(
// 		context.Background(),
// 		composeFileId,
// 		composeFileVersionRequestBody,
// 	)

// 	if err != nil {
// 		pterm.Printf("%s Failed to create compose file version\n", pterm.Red("❌"))
// 		return uuid.Nil, fmt.Errorf("failed to create compose file version: %w", err)
// 	}

// 	if composeFileVersionResponse.JSON200 != nil {
// 		return composeFileVersionResponse.JSON200.Id, nil
// 	}

// 	pterm.Printf("%s Failed to get compose file version ID\n", pterm.Red("❌"))
// 	return uuid.Nil, fmt.Errorf("failed to get compose file version ID")
// }

// func getTargetId(
// 	client *api.ClientWithResponses,
// 	organizationId uuid.UUID,
// 	appSlug string,
// 	appRequestBody api.UpsertAppV1JSONRequestBody,
// ) (uuid.UUID, error) {
// 	appResponse, err := client.UpsertAppV1WithResponse(
// 		context.Background(),
// 		organizationId,
// 		appSlug,
// 		appRequestBody,
// 	)

// 	if err != nil {
// 		pterm.Printf("%s Failed to upsert app\n", pterm.Red("❌"))
// 		return uuid.Nil, fmt.Errorf("failed to upsert app: %w", err)
// 	}

// 	if appResponse.JSON200 != nil {
// 		for _, environment := range *appResponse.JSON200.Environments {
// 			if environment.Targets != nil && environment.Slug == "production" {
// 				for _, target := range *environment.Targets {
// 					if target.Slug == "production" {
// 						return target.Id, nil
// 					}
// 				}
// 			}
// 		}
// 	}

// 	if appResponse.JSON201 != nil {
// 		for _, environment := range *appResponse.JSON201.Environments {
// 			if environment.Targets != nil && environment.Slug == "production" {
// 				for _, target := range *environment.Targets {
// 					if target.Slug == "production" {
// 						return target.Id, nil
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return uuid.Nil, fmt.Errorf("production target not found")
// }

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

	if version == "" {
		version, _ = determineVersion(yamlCompose)
		version, err = pterm.DefaultInteractiveTextInput.WithDefaultValue(version).Show("Enter version")
		if err != nil {
			return nil, fmt.Errorf("failed to get version input: %w", err)
		}

		if version == "" {
			return nil, fmt.Errorf("version cannot be empty")
		}
	}

	orgSlug, err := cfg.GetOrgSlug(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization slug: %w", err)
	}

	raw := string(yamlCompose)
	return client.CreateEnvironmentComposeFileWithResponse(
		context.Background(),
		orgSlug,
		cfg.GetAppSlug(),
		envName,
		api.CreateEnvironmentComposeFileJSONRequestBody{
			ComposeNoramlized: &composeConfigMap,
			ComposeRaw:        &raw,
			Version:           version,
		},
	)
}

func NewDeployCmd() *cobra.Command {
	var configPath string
	var envName string
	var version string

	cmd := &cobra.Command{
		Use:          "deploy",
		Short:        "Deploy a docker compose file",
		Long:         "Deploy a docker compose",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				err = initcmd.NewInitCmd().RunE(cmd, args)
				if err != nil {
					return fmt.Errorf("failed to initialize config: %w", err)
				}
				cfg, err = config.LoadConfig(configPath)
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
			}

			env, ok := cfg.Environments[envName]
			if !ok {
				// If no environment specified, use the first one
				if envName == "" {
					for name := range cfg.Environments {
						envName = name
						env = cfg.Environments[name]
						break
					}
				}
				if envName == "" {
					return fmt.Errorf("no environment specified and no environment found in config")
				}
			}

			composeFiles := env.ComposeFiles
			composeConfig, err := compose.LoadComposeConfig(composeFiles)
			if err != nil {
				pterm.Printf("%s Failed to load compose config\n", pterm.Red("❌"))
				return fmt.Errorf("failed to get services with build: %w", err)
			}

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
			fmt.Println()

			// // Building images
			// servicesWithBuild := composeConfig.GetServicesWithBuild()
			// if len(servicesWithBuild) > 0 {
			// 	cmd := fmt.Sprintf("docker compose -f %s build", composeFile)
			// 	pterm.Printf("🔨 Running: %s\n", pterm.Cyan(cmd))
			// 	if err := composeConfig.Build(); err != nil {
			// 		pterm.Printf("%s Build failed\n", pterm.Red("❌"))
			// 		return fmt.Errorf("failed to build images: %w", err)
			// 	}
			// 	pterm.Printf("✅ All images built successfully\n")
			// 	fmt.Println()
			// }

			// // Pulling external images
			// externalImages := composeConfig.GetExternalImages()
			// if len(externalImages) > 0 {
			// 	for _, image := range externalImages {
			// 		pterm.Printf("📥 Pulling: %s\n", pterm.Cyan(image))
			// 		if err := docker.PullImage(image); err != nil {
			// 			pterm.Printf("%s Failed to pull image: %s\n", pterm.Red("❌"), image)
			// 			return fmt.Errorf("failed to pull image %s: %w", image, err)
			// 		}
			// 	}
			// 	fmt.Println()
			// }

			// // Tagging images
			// images := composeConfig.GetImages()
			// if len(images) > 0 {
			// 	pterm.Println("ℹ️  Tagging Images")

			// 	for _, image := range images {
			// 		imageID, err := docker.GetImageID(image)
			// 		if err != nil {
			// 			pterm.Printf("%s Failed to get image ID for %s\n", pterm.Red("❌"), image)
			// 			return fmt.Errorf("failed to get image ID: %w", err)
			// 		}
			// 		imageName := strings.Split(image, ":")[0]
			// 		tag := fmt.Sprintf("%s/%s:%s", registry, imageName, imageID)
			// 		pterm.Printf("🏷️  Tagged: %s → %s\n", pterm.Cyan(image), pterm.Green(tag))
			// 	}
			// 	fmt.Println()
			// }

			client, err := api.NewViperClientWithResponses()
			if err != nil {
				pterm.Printf("%s Failed to create client\n", pterm.Red("❌"))
				return fmt.Errorf("failed to create client: %w", err)
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
				pterm.Printf("%s Failed to create compose file\n", pterm.Red("❌"))
				return fmt.Errorf("failed to create compose file")
			}

			pterm.Printf("✅ Created new version of compose file.\n")

			deployResponse, err := client.DeployEnvironmentComposeFileWithResponse(
				context.Background(),
				composeFileResponse.JSON200.Id,
			)

			if err != nil {
				return fmt.Errorf("failed to deploy environment compose file: %w", err)
			}

			if deployResponse.StatusCode() != 200 {
				pterm.Printf("%s Failed to deploy environment compose file\n", pterm.Red("❌"))
				return fmt.Errorf("failed to deploy environment compose file")
			}

			deployments := deployResponse.JSON200.Deployments

			if len(deployments) == 0 {
				pterm.Printf("No targets found to deploy to.\n")
				return nil
			}

			fmt.Println()
			spinner, _ := pterm.DefaultSpinner.Start("Waiting for deployments to complete...")
			// Wait for all deployments to complete
			for _, d := range deployments {
				for {
					deployment, err := client.GetDeploymentWithResponse(
						context.Background(),
						d.Id.String(),
					)
					if err != nil {
						return fmt.Errorf("failed to get deployment: %w", err)
					}

					if deployment.JSON200.Status == "failed" {
						return fmt.Errorf("deployment failed")
					}

					if deployment.JSON200.Status == "deployed" {
						break
					}

					// Sleep briefly before checking again
					time.Sleep(1 * time.Second)
				}
			}

			spinner.Success("Deployments completed")


			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", ".portway.yaml", "Config file to deploy")
	cmd.Flags().StringVarP(&envName, "env", "e", "", "Environment to deploy to")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version to deploy")

	return cmd
}
