package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AnthropicClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

const (
	anthropicTimeout         = 60 * time.Second
	anthropicMaxTokens       = 1024
	anthropicVersionHeader   = "2023-06-01"
	anthropicResponseMaxSize = 1 << 20
)

func NewAnthropicClient(apiKey string) *AnthropicClient {
	return NewAnthropicClientWithBaseURL(apiKey, "https://api.anthropic.com")
}

func NewAnthropicClientWithBaseURL(apiKey, baseURL string) *AnthropicClient {
	return &AnthropicClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: anthropicTimeout,
		},
	}
}

func (c *AnthropicClient) Complete(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	payload := struct {
		Model     string    `json:"model"`
		MaxTokens int       `json:"max_tokens"`
		System    string    `json:"system,omitempty"`
		Messages  []Message `json:"messages"`
	}{
		Model:     req.Model,
		MaxTokens: anthropicMaxTokens,
		System:    req.System,
		Messages:  []Message{},
	}

	payload.Messages = append(payload.Messages, req.Messages...)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersionHeader)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := readAllLimited(resp.Body, anthropicResponseMaxSize)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("anthropic error: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var decoded struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	for _, c := range decoded.Content {
		if c.Type == "text" && c.Text != "" {
			return &CompleteResponse{Text: c.Text}, nil
		}
	}
	return nil, fmt.Errorf("anthropic response missing text content")
}
