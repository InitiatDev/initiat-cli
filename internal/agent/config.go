package agent

import (
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/storage"
)

type RuntimeConfig struct {
	Enabled  bool
	Provider Provider
	Model    string
	APIKey   string
}

func LoadRuntimeConfig() (*RuntimeConfig, error) {
	cfg := config.Get()

	provider, err := NormalizeProvider(cfg.Agent.Provider)
	if err != nil {
		return nil, err
	}

	model := strings.TrimSpace(cfg.Agent.Model)
	if model == "" {
		return nil, fmt.Errorf("agent model not set (set with: initiat agent set-model <model>)")
	}

	store := storage.New()
	var apiKey string
	switch provider {
	case ProviderOpenAI:
		apiKey, err = store.GetAgentOpenAIAPIKey()
	case ProviderAnthropic:
		apiKey, err = store.GetAgentAnthropicAPIKey()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	if err != nil {
		return nil, err
	}

	return &RuntimeConfig{
		Enabled:  cfg.Agent.Enabled,
		Provider: provider,
		Model:    model,
		APIKey:   apiKey,
	}, nil
}

func NormalizeProvider(s string) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(ProviderOpenAI):
		return ProviderOpenAI, nil
	case string(ProviderAnthropic):
		return ProviderAnthropic, nil
	case "":
		return "", fmt.Errorf("agent provider not set (set with: initiat agent set-provider <openai|anthropic>)")
	default:
		return "", fmt.Errorf("invalid provider %q (expected openai or anthropic)", s)
	}
}
