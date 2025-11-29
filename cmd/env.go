package cmd

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/env"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments and secrets",
	Long:  `Manage local environments and sync secrets from Initiat Cloud.`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available environments",
	Long:  `List all cloud environments with their sync status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectCtx, err := GetProjectContext()
		if err != nil {
			return fmt.Errorf("failed to get project context: %w", err)
		}

		apiClient := client.New()
		environments, err := apiClient.ListEnvironments(projectCtx.OrgSlug, projectCtx.ProjectSlug)
		if err != nil {
			return fmt.Errorf("failed to list environments: %w", err)
		}

		if len(environments) == 0 {
			fmt.Println("No environments found.")
			return nil
		}

		activeEnv, _ := env.GetActiveEnvironment()

		fmt.Println("Environments:")
		for _, environment := range environments {
			marker := " "
			if environment.Slug == activeEnv {
				marker = "*"
			}

			fmt.Printf("%s %-10s (%d secrets)\n", marker, environment.Slug, environment.SecretsCount)
		}

		return nil
	},
}

var envSwitchCmd = &cobra.Command{
	Use:   "switch <slug>",
	Short: "Switch to an environment",
	Long:  `Switch to the specified environment and reload direnv.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		envSlug := args[0]

		projectCtx, err := GetProjectContext()
		if err != nil {
			return fmt.Errorf("failed to get project context: %w", err)
		}

		apiClient := client.New()
		_, err = apiClient.GetEnvironment(projectCtx.OrgSlug, projectCtx.ProjectSlug, envSlug)
		if err != nil {
			return fmt.Errorf("environment '%s' does not exist: %w", envSlug, err)
		}

		err = env.SetActiveEnvironment(envSlug)
		if err != nil {
			return fmt.Errorf("failed to set active environment: %w", err)
		}

		err = env.GenerateEnvrc()
		if err != nil {
			return fmt.Errorf("failed to generate .envrc: %w", err)
		}

		if env.CheckDirenvInstalled() {
			fmt.Printf("→ setting active -> environments/%s\n", envSlug)
			fmt.Printf("→ refreshing .envrc\n")
			fmt.Printf("→ direnv reload\n")

			err = env.ReloadDirenv()
			if err != nil {
				fmt.Printf("⚠️  direnv reload failed: %v\n", err)
				fmt.Printf("   Run 'direnv allow' to enable the environment\n")
			} else {
				fmt.Printf("Switched to \"%s\"\n", envSlug)
			}
		} else {
			fmt.Printf("⚠️  direnv not installed. Install with: %s\n", env.GetInstallInstructions())
			fmt.Printf("Switched to \"%s\" (run 'direnv allow' after installing direnv)\n", envSlug)
		}

		return nil
	},
}

var envCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current environment",
	Long:  `Show the currently active environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		activeEnv, err := env.GetActiveEnvironment()
		if err != nil {
			fmt.Println("No active environment")
			return nil
		}

		fmt.Printf("Current environment: %s\n", activeEnv)
		return nil
	},
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset the active environment",
	Long:  `Clear the currently active environment and reload direnv.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !env.IsInitCompleted() {
			return fmt.Errorf("initiat environment not initialized. Run 'initiat env init' first")
		}

		// Check if there's an active environment to unset
		activeEnv, err := env.GetActiveEnvironment()
		if err != nil {
			fmt.Println("No active environment to unset")
			return nil
		}

		err = env.UnsetActiveEnvironment()
		if err != nil {
			return fmt.Errorf("failed to unset active environment: %w", err)
		}

		fmt.Printf("Unset active environment: %s\n", activeEnv)

		if env.CheckDirenvInstalled() {
			fmt.Println("→ refreshing .envrc")
			fmt.Println("→ direnv reload")

			err = env.ReloadDirenv()
			if err != nil {
				fmt.Printf("⚠️  direnv reload failed: %v\n", err)
				fmt.Printf("   Run 'direnv allow' to enable the environment\n")
			} else {
				fmt.Println("Environment unset successfully")
			}
		} else {
			fmt.Printf("⚠️  direnv not installed. Install with: %s\n", env.GetInstallInstructions())
			fmt.Println("Environment unset (run 'direnv allow' after installing direnv)")
		}

		return nil
	},
}

var envLoadCmd = &cobra.Command{
	Use:    "load",
	Short:  "Load environment secrets (for direnv)",
	Long:   `Load and export secrets for the active environment. This command is intended to be used by direnv via eval.`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectCtx, err := GetProjectContext()
		if err != nil {
			return fmt.Errorf("failed to get project context: %w", err)
		}

		output, err := env.LoadEnvironmentSecrets(projectCtx.OrgSlug, projectCtx.ProjectSlug)
		if err != nil {
			return err
		}

		fmt.Print(output)
		return nil
	},
}

var envInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize environment setup",
	Long:  `Initialize the .initiat directory and check direnv setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := env.CreateInitiatDir()
		if err != nil {
			return fmt.Errorf("failed to create .initiat directory: %w", err)
		}

		fmt.Println("Created .initiat directory")

		gitStatus, err := env.GetGitignoreStatus()
		if err != nil {
			fmt.Printf("⚠️  Failed to check gitignore status: %v\n", err)
		} else {
			switch gitStatus {
			case "not a git repository":
				fmt.Println("ℹ️  Not a git repository - skipping .gitignore setup")
			case "missing":
				err = env.EnsureGitignore()
				if err != nil {
					fmt.Printf("⚠️  Failed to update .gitignore: %v\n", err)
				} else {
					fmt.Println("Updated .gitignore")
				}
			case "configured":
				fmt.Println(".gitignore already configured")
			}
		}

		if !env.CheckDirenvInstalled() {
			fmt.Printf("⚠️  direnv not installed. Install with: %s\n", env.GetInstallInstructions())
			return nil
		}

		version, err := env.GetDirenvVersion()
		if err != nil {
			fmt.Printf("⚠️  Failed to get direnv version: %v\n", err)
			return nil
		}

		fmt.Printf("direnv installed: %s\n", version)

		if !env.CheckDirenvHook() {
			fmt.Println("⚠️  direnv hook not found in shell configuration")
			fmt.Println("   Let me help you set it up...")

			shellType, err := env.PromptForShellType()
			if err != nil {
				return fmt.Errorf("failed to get shell type: %w", err)
			}

			err = env.WriteDirenvHookToShellConfig(shellType)
			if err != nil {
				return fmt.Errorf("failed to write direnv hook to shell config: %w", err)
			}

			fmt.Printf("✅ Added direnv hook to ~/.%src\n", shellType)
			fmt.Printf("   Please restart your shell or run 'source ~/.%src' to activate\n", shellType)
		} else {
			fmt.Println("direnv hook configured")
		}

		err = env.GenerateEnvrc()
		if err != nil {
			return fmt.Errorf("failed to create .envrc file: %w", err)
		}

		fmt.Println("Created .envrc file")

		if env.CheckDirenvInstalled() {
			fmt.Println("Running 'direnv allow'...")
			err = env.AllowDirenv()
			if err != nil {
				fmt.Printf("⚠️  direnv allow failed: %v\n", err)
				fmt.Println("   You may need to run 'direnv allow' manually")
			} else {
				fmt.Println("✅ direnv allow completed")
			}
		}

		return nil
	},
}

func init() {
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSwitchCmd)
	envCmd.AddCommand(envCurrentCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(envInitCmd)
	envCmd.AddCommand(envLoadCmd)

	rootCmd.AddCommand(envCmd)
}
