package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
	"github.com/InitiatDev/initiat-cli/internal/storage"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agent-assisted setup recovery",
	Long:  "Configure LLM provider settings and securely manage API keys for agent-assisted setup recovery.",
}

const (
	agentProviderOpenAI    = "openai"
	agentProviderAnthropic = "anthropic"
)

var agentEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable agent mode prompts on setup failure",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureConfigFileExists(); err != nil {
			return err
		}
		if err := config.Set("agent.enabled", true); err != nil {
			return err
		}
		if err := config.Save(); err != nil {
			return err
		}
		fmt.Println("✅ Agent mode enabled")
		return nil
	},
}

var agentDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable agent mode prompts on setup failure",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureConfigFileExists(); err != nil {
			return err
		}
		if err := config.Set("agent.enabled", false); err != nil {
			return err
		}
		if err := config.Save(); err != nil {
			return err
		}
		fmt.Println("✅ Agent mode disabled")
		return nil
	},
}

var agentSetProviderCmd = &cobra.Command{
	Use:   "set-provider <openai|anthropic>",
	Short: "Set the LLM provider for agent mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := normalizeAgentProvider(args[0])
		if err != nil {
			return err
		}
		if err := config.EnsureConfigFileExists(); err != nil {
			return err
		}
		if err := config.Set("agent.provider", provider); err != nil {
			return err
		}
		if err := config.Save(); err != nil {
			return err
		}
		fmt.Printf("✅ Agent provider set to %s\n", provider)
		return nil
	},
}

var agentSetModelCmd = &cobra.Command{
	Use:   "set-model <model>",
	Short: "Set the model name for agent mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		model := strings.TrimSpace(args[0])
		if model == "" {
			return fmt.Errorf("model cannot be empty")
		}
		if err := config.EnsureConfigFileExists(); err != nil {
			return err
		}
		if err := config.Set("agent.model", model); err != nil {
			return err
		}
		if err := config.Save(); err != nil {
			return err
		}
		fmt.Printf("✅ Agent model set to %s\n", model)
		return nil
	},
}

var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current agent configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		store := storage.New()

		openAIKeyStatus := "missing"
		if store.HasAgentOpenAIAPIKey() {
			openAIKeyStatus = "set"
		}

		anthropicKeyStatus := "missing"
		if store.HasAgentAnthropicAPIKey() {
			anthropicKeyStatus = "set"
		}

		fmt.Println("Agent configuration:")
		fmt.Printf("  enabled: %v\n", cfg.Agent.Enabled)
		if cfg.Agent.Provider == "" {
			fmt.Printf("  provider: (not set)\n")
		} else {
			fmt.Printf("  provider: %s\n", cfg.Agent.Provider)
		}
		if cfg.Agent.Model == "" {
			fmt.Printf("  model: (not set)\n")
		} else {
			fmt.Printf("  model: %s\n", cfg.Agent.Model)
		}
		fmt.Printf("  openai_api_key: %s\n", openAIKeyStatus)
		fmt.Printf("  anthropic_api_key: %s\n", anthropicKeyStatus)
		return nil
	},
}

var agentKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage agent provider API keys",
}

var agentKeySetCmd = &cobra.Command{
	Use:   "set <openai|anthropic>",
	Short: "Store a provider API key in the OS keyring",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := normalizeAgentProvider(args[0])
		if err != nil {
			return err
		}

		value, err := prompt.PromptHidden(fmt.Sprintf("%s API key", strings.ToUpper(provider)))
		if err != nil {
			return err
		}

		store := storage.New()
		switch provider {
		case agentProviderOpenAI:
			if err := store.StoreAgentOpenAIAPIKey(value); err != nil {
				return err
			}
		case agentProviderAnthropic:
			if err := store.StoreAgentAnthropicAPIKey(value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported provider: %s", provider)
		}

		fmt.Printf("✅ Stored %s API key in keyring\n", provider)
		return nil
	},
}

var agentKeyClearCmd = &cobra.Command{
	Use:   "clear <openai|anthropic>",
	Short: "Delete a provider API key from the OS keyring",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := normalizeAgentProvider(args[0])
		if err != nil {
			return err
		}

		store := storage.New()
		switch provider {
		case agentProviderOpenAI:
			_ = store.DeleteAgentOpenAIAPIKey()
		case agentProviderAnthropic:
			_ = store.DeleteAgentAnthropicAPIKey()
		default:
			return fmt.Errorf("unsupported provider: %s", provider)
		}

		fmt.Printf("✅ Cleared %s API key from keyring\n", provider)
		return nil
	},
}

func normalizeAgentProvider(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case agentProviderOpenAI:
		return agentProviderOpenAI, nil
	case agentProviderAnthropic:
		return agentProviderAnthropic, nil
	default:
		return "", fmt.Errorf("invalid provider %q (expected openai or anthropic)", s)
	}
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentEnableCmd)
	agentCmd.AddCommand(agentDisableCmd)
	agentCmd.AddCommand(agentSetProviderCmd)
	agentCmd.AddCommand(agentSetModelCmd)
	agentCmd.AddCommand(agentStatusCmd)
	agentCmd.AddCommand(agentKeyCmd)

	agentKeyCmd.AddCommand(agentKeySetCmd)
	agentKeyCmd.AddCommand(agentKeyClearCmd)
}
