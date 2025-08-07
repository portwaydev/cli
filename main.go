package main

import (
	"cli/cmd/auth"
	"cli/cmd/deploy"
	"cli/cmd/doctor"
	initcmd "cli/cmd/init"
	"cli/cmd/settings"
	"cli/cmd/validate"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/mitchellh/go-homedir"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "deploy",
		Short:        "Deploy a Docker Compose file to a Kubernetes cluster",
		Long:         `deploy is a CLI tool that deploys a Docker Compose file to a Kubernetes cluster. It supports deploying services, volumes, and networks to their Kubernetes equivalents.`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().String("token", "", "API key to use for authentication")

	rootCmd.AddCommand(deploy.NewDeployCmd())
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(auth.NewLoginCmd())
	rootCmd.AddCommand(settings.NewSettingsCmd())
	rootCmd.AddCommand(doctor.NewDoctorCmd())
	rootCmd.AddCommand(validate.NewValidateCmd())
	rootCmd.AddCommand(initcmd.NewInitCmd())

	return rootCmd
}

func init() {
	log.SetReportTimestamp(false)

	pterm.Success.Prefix.Text = "✅"
	pterm.Success.Prefix.Style = &pterm.ThemeDefault.SuccessMessageStyle
	pterm.Info.Prefix.Text = "ℹ️"
	pterm.Info.Prefix.Style = &pterm.ThemeDefault.InfoMessageStyle

	cobra.OnInitialize(initConfig)
}

func main() {
	rootCmd := NewRootCmd()

	viper.BindEnv("url", "PORTWAY_URL")
	viper.SetDefault("url", "https://portway.dev")

	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindEnv("token", "PORTWAY_API_KEY")
	viper.SetDefault("token", "")

	viper.SetDefault("autoupdate", true)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("Can't find home directory", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".portway")
		viper.SetConfigType("yaml")
		viper.SafeWriteConfig()
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config", err)
		os.Exit(1)
	}
}
