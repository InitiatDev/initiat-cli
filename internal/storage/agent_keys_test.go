package storage

import "testing"

func TestStorage_AgentAPIKeyOperations(t *testing.T) {
	s := NewWithKeyring("initiat-cli-test-agent-keys", NewMemKeyring())

	_ = s.DeleteAgentOpenAIAPIKey()
	_ = s.DeleteAgentAnthropicAPIKey()

	if s.HasAgentOpenAIAPIKey() {
		t.Fatalf("expected OpenAI key to be missing")
	}
	if s.HasAgentAnthropicAPIKey() {
		t.Fatalf("expected Anthropic key to be missing")
	}

	if err := s.StoreAgentOpenAIAPIKey("sk-openai-test"); err != nil {
		t.Fatalf("store OpenAI key: %v", err)
	}
	if err := s.StoreAgentAnthropicAPIKey("sk-anthropic-test"); err != nil {
		t.Fatalf("store Anthropic key: %v", err)
	}

	if !s.HasAgentOpenAIAPIKey() {
		t.Fatalf("expected OpenAI key to be set")
	}
	if !s.HasAgentAnthropicAPIKey() {
		t.Fatalf("expected Anthropic key to be set")
	}

	openAIKey, err := s.GetAgentOpenAIAPIKey()
	if err != nil {
		t.Fatalf("get OpenAI key: %v", err)
	}
	if openAIKey != "sk-openai-test" {
		t.Fatalf("unexpected OpenAI key: %q", openAIKey)
	}

	anthropicKey, err := s.GetAgentAnthropicAPIKey()
	if err != nil {
		t.Fatalf("get Anthropic key: %v", err)
	}
	if anthropicKey != "sk-anthropic-test" {
		t.Fatalf("unexpected Anthropic key: %q", anthropicKey)
	}

	if err := s.DeleteAgentOpenAIAPIKey(); err != nil {
		t.Fatalf("delete OpenAI key: %v", err)
	}
	if err := s.DeleteAgentAnthropicAPIKey(); err != nil {
		t.Fatalf("delete Anthropic key: %v", err)
	}

	if s.HasAgentOpenAIAPIKey() {
		t.Fatalf("expected OpenAI key to be missing after delete")
	}
	if s.HasAgentAnthropicAPIKey() {
		t.Fatalf("expected Anthropic key to be missing after delete")
	}
}
