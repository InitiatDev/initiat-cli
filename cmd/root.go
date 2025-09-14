package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/DylanBlakemore/initflow-cli/internal/config"
)

var (
	cfgFile string
	apiURL  string
)

var rootCmd = &cobra.Command{
	Use:   "initflow",
	Short: "InitFlow CLI",
	Long:  `InitFlow CLI — secure secrets, onboarding, and policy tooling.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		if apiURL != "" {
			if err := config.Set("api_base_url", apiURL); err != nil {
				return fmt.Errorf("failed to set API URL: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.initflow/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API base URL (default: https://api.initflow.com)")
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
