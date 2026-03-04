package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

const (
	openAITimeout         = 60 * time.Second
	openAITemperature     = 0.2
	openAIResponseMaxSize = 1 << 20
)

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return NewOpenAIClientWithBaseURL(apiKey, "https://api.openai.com")
}

func NewOpenAIClientWithBaseURL(apiKey, baseURL string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: openAITimeout,
		},
	}
}

func (c *OpenAIClient) Complete(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	payload := struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Temperature float64   `json:"temperature"`
	}{
		Model:       req.Model,
		Messages:    []Message{},
		Temperature: openAITemperature,
	}

	if req.System != "" {
		payload.Messages = append(payload.Messages, Message{Role: "system", Content: req.System})
	}
	payload.Messages = append(payload.Messages, req.Messages...)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := readAllLimited(resp.Body, openAIResponseMaxSize)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai error: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return nil, fmt.Errorf("openai response missing choices")
	}

	return &CompleteResponse{Text: decoded.Choices[0].Message.Content}, nil
}
