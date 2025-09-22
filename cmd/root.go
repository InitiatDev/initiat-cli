package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/DylanBlakemore/initiat-cli/internal/config"
)

var (
	cfgFile     string
	apiURL      string
	serviceName string
)

var rootCmd = &cobra.Command{
	Use:   "initiat",
	Short: "Initiat CLI",
	Long:  `Initiat CLI — secure secrets, onboarding, and policy tooling.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		if apiURL != "" {
			if err := config.Set("api_base_url", apiURL); err != nil {
				return fmt.Errorf("failed to set API URL: %w", err)
			}
		}

		if serviceName != "" {
			if err := config.Set("service_name", serviceName); err != nil {
				return fmt.Errorf("failed to set service name: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.initiat/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API base URL (default: https://api.initiat.com)")
	rootCmd.PersistentFlags().StringVar(&serviceName, "service-name", "initiat-cli",
		"keyring service name for credential storage")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
