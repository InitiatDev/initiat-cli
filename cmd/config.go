package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DylanBlakemore/initflow-cli/internal/config"
	"github.com/DylanBlakemore/initflow-cli/internal/slug"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Manage CLI configuration including default organization context.`,
}

var configUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Set configuration values",
	Long:  `Set configuration values for the CLI.`,
}

var configUseOrgCmd = &cobra.Command{
	Use:   "org <org-slug>",
	Short: "Set default organization",
	Long: `Set the default organization for workspace operations.

When a default organization is set, you can use workspace commands with just the workspace slug:
  initflow workspace init production
  initflow secret list staging

Examples:
  initflow config use org acme-corp
  initflow config use org my-company`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigUseOrg,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Show the current CLI configuration including default organization.`,
	RunE:  runConfigShow,
}

var configClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear configuration values",
	Long:  `Clear configuration values.`,
}

var configClearOrgCmd = &cobra.Command{
	Use:   "org",
	Short: "Clear default organization",
	Long:  `Clear the default organization setting.`,
	RunE:  runConfigClearOrg,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configClearCmd)

	configUseCmd.AddCommand(configUseOrgCmd)
	configClearCmd.AddCommand(configClearOrgCmd)
}

func runConfigUseOrg(cmd *cobra.Command, args []string) error {
	orgSlug := args[0]

	if err := slug.ValidateSlug(orgSlug); err != nil {
		return fmt.Errorf("❌ Invalid organization slug: %w", err)
	}

	if err := config.SetDefaultOrgSlug(orgSlug); err != nil {
		return fmt.Errorf("❌ Failed to set default organization: %w", err)
	}

	fmt.Printf("✅ Default organization set to '%s'\n", orgSlug)
	fmt.Println("💡 You can now use workspace commands with just the workspace slug:")
	fmt.Printf("   initflow workspace init <workspace-slug>\n")
	fmt.Printf("   initflow secret list <workspace-slug>\n")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	fmt.Println("Current configuration:")
	fmt.Printf("  API Base URL: %s\n", cfg.APIBaseURL)

	if cfg.DefaultOrgSlug != "" {
		fmt.Printf("  Default Organization: %s\n", cfg.DefaultOrgSlug)
	} else {
		fmt.Println("  Default Organization: (not set)")
		fmt.Println("💡 Set a default organization with: initflow config use org <org-slug>")
	}

	return nil
}

func runConfigClearOrg(cmd *cobra.Command, args []string) error {
	if err := config.ClearDefaultOrgSlug(); err != nil {
		return fmt.Errorf("❌ Failed to clear default organization: %w", err)
	}

	fmt.Println("✅ Default organization cleared")
	fmt.Println("💡 You'll need to use full composite slugs: org-slug/workspace-slug")

	return nil
}
