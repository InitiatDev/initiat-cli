package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type ToolRunner interface {
	RunCommand(ctx context.Context, action ProposedAction) error
	EditFiles(ctx context.Context, action ProposedAction) error
}

type Orchestrator struct {
	llm      LLM
	model    string
	approver Approver
	tools    ToolRunner
}

// NOTE: We intentionally use a simple "return JSON" prompt here rather than provider-native tool calling.
// Alternative (more complex, more robust): define a propose_actions tool with a strict JSON Schema and force
// a tool call (OpenAI function calling + Anthropic tool_use), then parse tool-call arguments instead of model text.

func NewOrchestrator(llm LLM, model string, approver Approver, tools ToolRunner) (*Orchestrator, error) {
	if llm == nil {
		return nil, fmt.Errorf("llm is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model is required")
	}
	if approver == nil {
		return nil, fmt.Errorf("approver is required")
	}
	return &Orchestrator{
		llm:      llm,
		model:    model,
		approver: approver,
		tools:    tools,
	}, nil
}

func (o *Orchestrator) Diagnose(ctx context.Context, report *setup.ExecutionReport) (*Decision, error) {
	if report == nil {
		return nil, fmt.Errorf("report is nil")
	}

	payload, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}

	system := `You are a senior engineer diagnosing a failing developer setup run.

Prefer safe, read-only diagnosis commands before file edits. Propose 1-3 actions at a time.
Never request or output secrets. Assume the report is already redacted.
Avoid destructive actions unless absolutely necessary and clearly justified.

Return ONLY valid JSON matching this schema:
{
  "explanation": "string",
  "actions": [
    {
      "type": "stop" | "ask_user" | "run_command" | "edit_files",
      "reason": "string",
      "command": "string (if run_command)",
      "cwd": "string (optional)",
      "env": {"K":"V"} (optional),
      "prompt": "string (if ask_user)",
      "edits": [{"path":"string","content":"string"}] (if edit_files)
    }
  ]
}`

	user := "Here is the failing setup execution report JSON:\n" + string(payload)
	resp, err := o.llm.Complete(ctx, &CompleteRequest{
		Model:  o.model,
		System: system,
		Messages: []Message{
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return nil, err
	}

	decision, err := ParseDecision(resp.Text)
	if err != nil {
		return nil, err
	}
	return decision, nil
}

func (o *Orchestrator) Apply(ctx context.Context, decision *Decision) error {
	if decision == nil {
		return fmt.Errorf("decision is nil")
	}

	for _, action := range decision.Actions {
		ok, err := o.approver.Approve(action)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("action not approved")
		}

		switch action.Type {
		case ActionAskUser, ActionStop:
			continue
		case ActionRunCommand:
			if o.tools == nil {
				return fmt.Errorf("no tool runner configured")
			}
			if err := o.tools.RunCommand(ctx, action); err != nil {
				return err
			}
		case ActionEditFiles:
			if o.tools == nil {
				return fmt.Errorf("no tool runner configured")
			}
			if err := o.tools.EditFiles(ctx, action); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown action type: %s", action.Type)
		}
	}

	return nil
}

func ParseDecision(s string) (*Decision, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return nil, fmt.Errorf("empty decision")
	}

	var d Decision
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return nil, fmt.Errorf("invalid decision json: %w", err)
	}
	if strings.TrimSpace(d.Explanation) == "" {
		return nil, fmt.Errorf("decision missing explanation")
	}
	if len(d.Actions) == 0 {
		return nil, fmt.Errorf("decision missing actions")
	}
	for i := range d.Actions {
		if d.Actions[i].Type == "" {
			return nil, fmt.Errorf("action[%d] missing type", i)
		}
		switch d.Actions[i].Type {
		case ActionStop:
		case ActionAskUser:
			if strings.TrimSpace(d.Actions[i].Prompt) == "" {
				return nil, fmt.Errorf("action[%d] ask_user missing prompt", i)
			}
		case ActionRunCommand:
			if strings.TrimSpace(d.Actions[i].Command) == "" {
				return nil, fmt.Errorf("action[%d] run_command missing command", i)
			}
		case ActionEditFiles:
			if len(d.Actions[i].Edits) == 0 {
				return nil, fmt.Errorf("action[%d] edit_files missing edits", i)
			}
		default:
			return nil, fmt.Errorf("action[%d] has unknown type %q", i, d.Actions[i].Type)
		}
	}
	return &d, nil
}
