package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type ToolRunner interface {
	RunCommand(ctx context.Context, action ProposedAction) (string, error)
	EditFiles(ctx context.Context, action ProposedAction) error
	ListFiles(ctx context.Context, action ProposedAction) (string, error)
	ReadFiles(ctx context.Context, action ProposedAction) (string, error)
}

type Orchestrator struct {
	llm      LLM
	model    string
	approver Approver
	tools    ToolRunner

	promptInput func(prompt string) (string, error)
	debugWriter io.Writer
}

const (
	maxInterrogationSteps = 20
	maxReadFilesPerAction = 6
)

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

func (o *Orchestrator) SetDebugWriter(w io.Writer) *Orchestrator {
	o.debugWriter = w
	return o
}

func (o *Orchestrator) SetPromptInput(fn func(prompt string) (string, error)) *Orchestrator {
	o.promptInput = fn
	return o
}

func (o *Orchestrator) Diagnose(ctx context.Context, report *setup.ExecutionReport) (*Decision, error) {
	return o.DiagnoseWithContext(ctx, report, "")
}

func (o *Orchestrator) DiagnoseWithContext(
	ctx context.Context,
	report *setup.ExecutionReport,
	extraContext string,
) (*Decision, error) {
	if report == nil {
		return nil, fmt.Errorf("report is nil")
	}

	payload, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}

	system := buildDiagnoseSystemPrompt()
	messages := buildDiagnoseMessages(payload, extraContext)
	return o.runDiagnoseLoop(ctx, system, messages)
}

func (o *Orchestrator) Apply(ctx context.Context, decision *Decision) error {
	_, err := o.ApplyWithResults(ctx, decision)
	return err
}

func (o *Orchestrator) ApplyWithResults(ctx context.Context, decision *Decision) (*ApplyResult, error) {
	if decision == nil {
		return nil, fmt.Errorf("decision is nil")
	}

	out := &ApplyResult{}
	for _, action := range decision.Actions {
		ok, err := o.approver.Approve(action)
		if err != nil {
			return out, err
		}
		if !ok {
			return out, fmt.Errorf(
				"action not approved: type=%s summary=%q",
				action.Type,
				action.Summary,
			)
		}

		r := AppliedActionResult{
			Type:    action.Type,
			Summary: action.Summary,
			OK:      true,
		}
		if err := o.applyOne(ctx, action, &r); err != nil {
			return out, err
		}
		o.logApplied(r)
		out.Results = append(out.Results, r)
	}

	return out, nil
}

func (o *Orchestrator) applyOne(ctx context.Context, action ProposedAction, r *AppliedActionResult) error {
	switch action.Type {
	case ActionAskUser:
		if o.promptInput != nil {
			resp, err := o.promptInput(action.Prompt)
			if err != nil {
				r.OK = false
				r.Error = err.Error()
			} else {
				r.Output = resp
			}
		}
		return nil
	case ActionStop:
		return nil
	case ActionListFiles:
		if o.tools == nil {
			return fmt.Errorf("no tool runner configured")
		}
		out, err := o.tools.ListFiles(ctx, action)
		if err != nil {
			r.OK = false
			r.Error = err.Error()
		} else {
			r.Output = out
		}
		return nil
	case ActionReadFiles:
		if o.tools == nil {
			return fmt.Errorf("no tool runner configured")
		}
		out, err := o.tools.ReadFiles(ctx, action)
		if err != nil {
			r.OK = false
			r.Error = err.Error()
		} else {
			r.Output = out
		}
		return nil
	case ActionRunCommand:
		if o.tools == nil {
			return fmt.Errorf("no tool runner configured")
		}
		var cmdOut string
		if err := callSafely(func() error {
			out, err := o.tools.RunCommand(ctx, action)
			cmdOut = out
			return err
		}); err != nil {
			r.OK = false
			r.Error = err.Error()
			r.Output = cmdOut // stderr on failure
		} else {
			r.Output = cmdOut // stdout on success
		}
		return nil
	case ActionEditFiles:
		if o.tools == nil {
			return fmt.Errorf("no tool runner configured")
		}
		if err := callSafely(func() error { return o.tools.EditFiles(ctx, action) }); err != nil {
			r.OK = false
			r.Error = err.Error()
		}
		return nil
	}
	return fmt.Errorf("unknown action type: %s", action.Type)
}

func callSafely(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn()
}

func (o *Orchestrator) logApplied(r AppliedActionResult) {
	if r.OK {
		return
	}
	o.logln(
		fmt.Sprintf(
			"[agent] action failed: %s summary=%q err=%s\n",
			r.Type,
			r.Summary,
			r.Error,
		),
	)
}

func buildDiagnoseSystemPrompt() string {
	return `You are a senior engineer diagnosing a failing developer setup run.

Prefer safe, read-only workspace inspection before shell commands or file edits.
Propose 1-3 actions at a time.
Never request or output secrets. Assume the report is already redacted.
Avoid destructive actions unless absolutely necessary and clearly justified.

For list_files/read_files: you MUST only reference paths within the project base directory
(no absolute paths like /Users/...).

If a README is present (see project snapshot), your FIRST step should be to read it via read_files.
Bias strongly towards reading only: README + important_files from the project snapshot.
Avoid reading arbitrary docs (AGENTS.md, CONTRIBUTING.md, etc) unless absolutely necessary;
if needed, ask_user with a specific justification.
Keep read_files small: at most a handful of files per action.

Assume .initiat/setup.yml may be incorrect or out-of-date.
Prefer aligning it to the README and observed project structure over treating it as ground truth.
When the failing step is a setup.yml command (e.g. db:migrate), prefer editing setup.yml
to use the README-recommended workflow before running ad-hoc introspection commands.
If a previous action failed, do NOT repeat it; instead change the approach (usually: edit setup.yml).

Return ONLY valid JSON matching this schema:
{
  "explanation": "string",
  "actions": [
    {
      "type": "stop" | "ask_user" | "list_files" | "read_files" | "run_command" | "edit_files",
      "summary": "string (plain English)",
      "danger": "safe" | "caution" | "dangerous",
      "danger_reason": "string",
      "reason": "string",
      "path": "string (optional; if list_files)",
      "paths": ["string"] (if read_files),
      "command": "string (if run_command)",
      "cwd": "string (optional)",
      "env": {"K":"V"} (optional),
      "prompt": "string (if ask_user)",
      "edits": [{"path":"string","content":"string"}] (if edit_files)
    }
  ]
}`
}

func buildDiagnoseMessages(payload []byte, extraContext string) []Message {
	extraContext = strings.TrimSpace(extraContext)
	msgs := []Message{
		{Role: "user", Content: "Here is the failing setup execution report JSON:\n" + string(payload)},
	}
	if extraContext != "" {
		msgs = append(msgs, Message{Role: "user", Content: extraContext})
	}
	return msgs
}

func (o *Orchestrator) runDiagnoseLoop(
	ctx context.Context,
	system string,
	messages []Message,
) (*Decision, error) {
	var lastDecision *Decision

	for step := 0; step < maxInterrogationSteps; step++ {
		resp, err := o.llm.Complete(ctx, &CompleteRequest{
			Model:    o.model,
			System:   system,
			Messages: messages,
		})
		if err != nil {
			return nil, err
		}

		decision, err := ParseDecision(resp.Text)
		if err != nil {
			return nil, err
		}

		lastDecision = decision
		o.logDecision(step, decision)

		if !decisionHasInterrogation(decision) {
			return decision, nil
		}
		if o.tools == nil {
			return nil, fmt.Errorf("decision requested interrogation but no tool runner configured")
		}

		results := o.applyInterrogationActions(ctx, decision)
		if strings.TrimSpace(results) == "" {
			results = "(no results)"
		}

		messages = append(messages, Message{
			Role: "user",
			Content: "Workspace interrogation results (read-only):\n" + results +
				"\n\nBased on these results, propose next actions (prefer smallest safe step).",
		})
	}

	if lastDecision != nil {
		return nil, fmt.Errorf(
			"too many interrogation steps without actionable decision (last_explanation=%q)",
			lastDecision.Explanation,
		)
	}
	return nil, fmt.Errorf("too many interrogation steps without actionable decision")
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
		if err := validateAction(i, d.Actions[i]); err != nil {
			return nil, err
		}
	}
	return &d, nil
}

func validateAction(i int, a ProposedAction) error {
	if a.Type == "" {
		return fmt.Errorf("action[%d] missing type", i)
	}
	if strings.TrimSpace(a.Summary) == "" {
		return fmt.Errorf("action[%d] missing summary", i)
	}
	switch a.Danger {
	case DangerSafe, DangerCaution, DangerDangerous:
	default:
		return fmt.Errorf("action[%d] has invalid danger %q", i, a.Danger)
	}
	if strings.TrimSpace(a.DangerReason) == "" {
		return fmt.Errorf("action[%d] missing danger_reason", i)
	}

	switch a.Type {
	case ActionStop:
		return nil
	case ActionAskUser:
		if strings.TrimSpace(a.Prompt) == "" {
			return fmt.Errorf("action[%d] ask_user missing prompt", i)
		}
		return nil
	case ActionListFiles:
		return nil
	case ActionReadFiles:
		if len(a.Paths) == 0 {
			return fmt.Errorf("action[%d] read_files missing paths", i)
		}
		if len(a.Paths) > maxReadFilesPerAction {
			return fmt.Errorf(
				"action[%d] read_files has too many paths (max=%d)",
				i,
				maxReadFilesPerAction,
			)
		}
		return nil
	case ActionRunCommand:
		if strings.TrimSpace(a.Command) == "" {
			return fmt.Errorf("action[%d] run_command missing command", i)
		}
		return nil
	case ActionEditFiles:
		if len(a.Edits) == 0 {
			return fmt.Errorf("action[%d] edit_files missing edits", i)
		}
		return nil
	}

	return fmt.Errorf("action[%d] has unknown type %q", i, a.Type)
}

func decisionHasInterrogation(d *Decision) bool {
	if d == nil {
		return false
	}
	for i := range d.Actions {
		if d.Actions[i].Type == ActionListFiles || d.Actions[i].Type == ActionReadFiles {
			return true
		}
	}
	return false
}

func (o *Orchestrator) applyInterrogationActions(ctx context.Context, d *Decision) string {
	var b strings.Builder
	for i := range d.Actions {
		a := d.Actions[i]
		if a.Type == ActionListFiles {
			out, err := o.tools.ListFiles(ctx, a)
			if b.Len() > 0 {
				b.WriteString("\n\n")
			}
			pathLabel := a.Path
			if strings.TrimSpace(pathLabel) == "" {
				pathLabel = "."
			}
			b.WriteString("==> list_files ")
			b.WriteString(pathLabel)
			b.WriteString(" <==\n")
			if err != nil {
				b.WriteString("ERROR: ")
				b.WriteString(err.Error())
				continue
			}
			b.WriteString(out)
			continue
		}
		if a.Type == ActionReadFiles {
			out, err := o.tools.ReadFiles(ctx, a)
			if b.Len() > 0 {
				b.WriteString("\n\n")
			}
			if err != nil {
				b.WriteString("==> read_files <==\n")
				b.WriteString("ERROR: ")
				b.WriteString(err.Error())
				continue
			}
			b.WriteString(out)
		}
	}
	return b.String()
}

func (o *Orchestrator) logln(s string) {
	if o == nil || o.debugWriter == nil {
		return
	}
	_, _ = io.WriteString(o.debugWriter, s)
}

func (o *Orchestrator) logDecision(step int, d *Decision) {
	if d == nil {
		return
	}

	o.logln("\n[agent] step " + fmt.Sprintf("%d", step+1) + "/" + fmt.Sprintf("%d", maxInterrogationSteps) + "\n")
	o.logln("[agent] explanation:\n" + d.Explanation + "\n")
	if len(d.Actions) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString("[agent] planned actions:\n")
	for _, a := range d.Actions {
		b.WriteString("- ")
		b.WriteString(string(a.Type))
		b.WriteString(" (")
		b.WriteString(string(a.Danger))
		b.WriteString("): ")
		b.WriteString(a.Summary)
		b.WriteString("\n")
		switch a.Type {
		case ActionListFiles:
			p := a.Path
			if strings.TrimSpace(p) == "" {
				p = "."
			}
			b.WriteString("  - path: ")
			b.WriteString(p)
			b.WriteString("\n")
		case ActionReadFiles:
			b.WriteString("  - paths: ")
			b.WriteString(strings.Join(a.Paths, ", "))
			b.WriteString("\n")
		case ActionRunCommand:
			b.WriteString("  - command: ")
			b.WriteString(a.Command)
			b.WriteString("\n")
		case ActionEditFiles:
			b.WriteString("  - edits: ")
			b.WriteString(fmt.Sprintf("%d file(s)", len(a.Edits)))
			b.WriteString("\n")
		case ActionStop, ActionAskUser:
		}
	}
	o.logln(b.String())
}
