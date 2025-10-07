package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/InitiatDev/initiat-cli/internal/config"
)

const (
	configSetArgsCount       = 2
	projectPathPartsExpected = 2
	yesResponse              = "yes"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Manage CLI configuration including API settings, project defaults, and aliases.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value using dot notation for nested keys.

Examples:
  initiat config set api.url "https://www.initiat.dev"
  initiat config set api.timeout "60s"
  initiat config set org "my-company"
  initiat config set project "production"
  initiat config set service "my-custom-service"`,
	Args: cobra.ExactArgs(configSetArgsCount),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value using dot notation for nested keys.

Examples:
  initiat config get api.url
  initiat config get org
  initiat config get project`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration",
	Long:  `Show all current CLI configuration values.`,
	RunE:  runConfigShow,
}

var configClearCmd = &cobra.Command{
	Use:   "clear <key>",
	Short: "Clear a configuration value",
	Long: `Clear a configuration value using dot notation for nested keys.

Examples:
  initiat config clear org
  initiat config clear api.timeout`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigClear,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long: `Reset all configuration values to their default settings.

This will:
- Reset all API settings to defaults
- Clear project defaults (org and project)
- Remove all project aliases
- Reset service name to default

Examples:
  initiat config reset`,
	RunE: runConfigReset,
}

var configAliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage project aliases",
	Long:  `Manage project aliases for convenient project references.`,
}

var configAliasSetCmd = &cobra.Command{
	Use:   "set <alias> <project-path>",
	Short: "Set a project alias",
	Long: `Set a project alias to a full project path.

Examples:
  initiat config alias set prod "acme-corp/production"
  initiat config alias set staging "acme-corp/staging"
  initiat config alias set dev "acme-corp/development"`,
	Args: cobra.ExactArgs(configSetArgsCount),
	RunE: runConfigAliasSet,
}

var configAliasGetCmd = &cobra.Command{
	Use:   "get <alias>",
	Short: "Get a project alias",
	Long: `Get the project path for a specific alias.

Examples:
  initiat config alias get prod`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigAliasGet,
}

var configAliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all project aliases",
	Long:  `List all configured project aliases.`,
	RunE:  runConfigAliasList,
}

var configAliasRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove a project alias",
	Long: `Remove a project alias.

Examples:
  initiat config alias remove prod`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigAliasRemove,
}

var (
	clearAll bool
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configClearCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configAliasCmd)

	configAliasCmd.AddCommand(configAliasSetCmd)
	configAliasCmd.AddCommand(configAliasGetCmd)
	configAliasCmd.AddCommand(configAliasListCmd)
	configAliasCmd.AddCommand(configAliasRemoveCmd)

	configClearCmd.Flags().BoolVar(&clearAll, "all", false, "Clear all configuration values")
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	if !config.IsValidConfigKey(key) {
		return fmt.Errorf("❌ Unknown configuration key: %s\nValid keys: %s",
			key, strings.Join(config.GetAllConfigKeys(), ", "))
	}

	if err := config.EnsureConfigFileExists(); err != nil {
		return err
	}

	actualKey := config.MapSimplifiedKey(key)
	if err := config.Set(actualKey, value); err != nil {
		return fmt.Errorf("❌ Failed to set configuration: %w", err)
	}

	if err := config.Save(); err != nil {
		return fmt.Errorf("❌ Failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Set %s = %s\n", key, value)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	if !config.IsValidConfigKey(key) {
		return fmt.Errorf("❌ Unknown configuration key: %s\nValid keys: %s",
			key, strings.Join(config.GetAllConfigKeys(), ", "))
	}

	actualKey := config.MapSimplifiedKey(key)
	value := viper.Get(actualKey)

	if str, ok := value.(string); ok && str == "" {
		fmt.Printf("%s: (not set)\n", key)
	} else {
		fmt.Printf("%s: %v\n", key, value)
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	fmt.Println("Current configuration:")
	fmt.Printf("  api.url: %s\n", cfg.API.BaseURL)
	fmt.Printf("  api.timeout: %s\n", cfg.API.Timeout)
	fmt.Printf("  service: %s\n", cfg.ServiceName)

	if cfg.Project.DefaultOrg != "" {
		fmt.Printf("  org: %s\n", cfg.Project.DefaultOrg)
	} else {
		fmt.Printf("  org: (not set)\n")
	}

	if cfg.Project.DefaultProject != "" {
		fmt.Printf("  project: %s\n", cfg.Project.DefaultProject)
	} else {
		fmt.Printf("  project: (not set)\n")
	}

	aliases := config.ListAliases()
	if len(aliases) > 0 {
		fmt.Println("\nProject aliases:")
		for alias, path := range aliases {
			fmt.Printf("  %s: %s\n", alias, path)
		}
	} else {
		fmt.Println("\nProject aliases: (none configured)")
	}

	return nil
}

func runConfigClear(cmd *cobra.Command, args []string) error {
	if clearAll {
		return runConfigClearAll()
	}

	key := args[0]

	if err := config.EnsureConfigFileExists(); err != nil {
		return err
	}

	actualKey := config.MapSimplifiedKey(key)
	if err := config.Set(actualKey, ""); err != nil {
		return fmt.Errorf("❌ Failed to clear configuration: %w", err)
	}

	if err := config.Save(); err != nil {
		return fmt.Errorf("❌ Failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Cleared %s\n", key)
	return nil
}

func runConfigClearAll() error {
	if err := config.EnsureConfigFileExists(); err != nil {
		return err
	}

	fmt.Print("⚠️  Are you sure you want to clear all configuration? (y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != yesResponse {
		fmt.Println("❌ Clear all cancelled")
		return nil
	}

	for _, key := range config.ConfigKeys {
		if err := config.Set(key.Actual, key.Default); err != nil {
			return fmt.Errorf("❌ Failed to reset %s: %w", key.Simplified, err)
		}
	}

	if err := config.Set("aliases", make(map[string]string)); err != nil {
		return fmt.Errorf("❌ Failed to reset aliases: %w", err)
	}

	if err := config.Save(); err != nil {
		return fmt.Errorf("❌ Failed to save configuration: %w", err)
	}

	fmt.Println("✅ All configuration cleared and reset to defaults")
	return nil
}

func runConfigAliasSet(cmd *cobra.Command, args []string) error {
	alias := args[0]
	projectPath := args[1]

	if !strings.Contains(projectPath, "/") {
		return fmt.Errorf("❌ Project path must be in format 'org/project', got: %s", projectPath)
	}

	parts := strings.Split(projectPath, "/")
	if len(parts) != projectPathPartsExpected {
		return fmt.Errorf("❌ Project path must be in format 'org/project', got: %s", projectPath)
	}

	if err := config.EnsureConfigFileExists(); err != nil {
		return err
	}

	if err := config.SetAlias(alias, projectPath); err != nil {
		return fmt.Errorf("❌ Failed to set alias: %w", err)
	}

	fmt.Printf("✅ Set alias '%s' = %s\n", alias, projectPath)
	return nil
}

func runConfigAliasGet(cmd *cobra.Command, args []string) error {
	alias := args[0]
	path := config.GetAlias(alias)

	if path == "" {
		fmt.Printf("❌ Alias '%s' not found\n", alias)
		return nil
	}

	fmt.Printf("%s: %s\n", alias, path)
	return nil
}

func runConfigAliasList(cmd *cobra.Command, args []string) error {
	aliases := config.ListAliases()

	if len(aliases) == 0 {
		fmt.Println("No project aliases configured")
		fmt.Println("💡 Set an alias with: initiat config alias set <alias> <org/project>")
		return nil
	}

	fmt.Println("Project aliases:")
	for alias, path := range aliases {
		fmt.Printf("  %s: %s\n", alias, path)
	}

	return nil
}

func runConfigAliasRemove(cmd *cobra.Command, args []string) error {
	alias := args[0]

	if err := config.EnsureConfigFileExists(); err != nil {
		return err
	}

	if err := config.RemoveAlias(alias); err != nil {
		return fmt.Errorf("❌ Failed to remove alias: %w", err)
	}

	fmt.Printf("✅ Removed alias '%s'\n", alias)
	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	fmt.Print("⚠️  Are you sure you want to reset all configuration to defaults? (y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != yesResponse {
		fmt.Println("❌ Reset cancelled")
		return nil
	}

	if err := config.ResetToDefaults(); err != nil {
		return fmt.Errorf("❌ Failed to reset configuration: %w", err)
	}

	fmt.Println("✅ Configuration reset to defaults")
	return nil
}
