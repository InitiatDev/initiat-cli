package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type IssueBuckets struct {
	SetupOrApp       []string `json:"setup_or_app"`
	LocalEnvironment []string `json:"local_environment"`
	Notes            string   `json:"notes,omitempty"`
}

func AssessIssues(
	ctx context.Context,
	llm LLM,
	model string,
	report *setup.ExecutionReport,
	decision *Decision,
) (*IssueBuckets, error) {
	if llm == nil {
		return nil, fmt.Errorf("llm is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model is required")
	}
	if report == nil {
		return nil, fmt.Errorf("report is nil")
	}
	if decision == nil {
		return nil, fmt.Errorf("decision is nil")
	}

	system := `You are classifying problems encountered during a developer setup run.
Return ONLY valid JSON with this schema:
{
  "setup_or_app": ["string", ...],
  "local_environment": ["string", ...],
  "notes": "string (optional)"
}

Rules:
- setup_or_app: issues in the setup script, repo, dependencies, or the target app itself that should affect most users.
- local_environment: issues unique to this machine/user environment (OS quirks, missing global tools,
  local permissions, etc.).
- Keep each item short and actionable. Prefer 1-5 items per bucket.
- Never include secrets.`

	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		return nil, fmt.Errorf("marshal decision: %w", err)
	}

	user := "Setup execution report JSON:\n" + string(reportJSON) + "\n\nAgent decision JSON:\n" + string(decisionJSON)
	resp, err := llm.Complete(ctx, &CompleteRequest{
		Model:  model,
		System: system,
		Messages: []Message{
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(resp.Text)
	if raw == "" {
		return nil, fmt.Errorf("empty issue assessment")
	}

	var out IssueBuckets
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("invalid issue assessment json: %w", err)
	}
	if out.SetupOrApp == nil {
		out.SetupOrApp = []string{}
	}
	if out.LocalEnvironment == nil {
		out.LocalEnvironment = []string{}
	}
	return &out, nil
}
