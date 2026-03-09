package agent

import (
	"context"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type mockLLM struct {
	text string
	err  error
}

func (m *mockLLM) Complete(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CompleteResponse{Text: m.text}, nil
}

type seqLLM struct {
	texts []string
}

func (s *seqLLM) Complete(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error) {
	if len(s.texts) == 0 {
		return &CompleteResponse{Text: `{"explanation":"x","actions":[{"type":"stop","summary":"s","danger":"safe","danger_reason":"dr","reason":"done"}]}`}, nil
	}
	out := s.texts[0]
	s.texts = s.texts[1:]
	return &CompleteResponse{Text: out}, nil
}

type mockApprover struct {
	ok bool
}

func (a *mockApprover) Approve(action ProposedAction) (bool, error) {
	return a.ok, nil
}

type mockTools struct {
	ran   int
	edits int
}

func (t *mockTools) RunCommand(ctx context.Context, action ProposedAction) error {
	t.ran++
	return nil
}

func (t *mockTools) EditFiles(ctx context.Context, action ProposedAction) error {
	t.edits++
	return nil
}

func (t *mockTools) ListFiles(ctx context.Context, action ProposedAction) (string, error) {
	return "", nil
}

func (t *mockTools) ReadFiles(ctx context.Context, action ProposedAction) (string, error) {
	return "", nil
}

func TestParseDecision_ValidRunCommand(t *testing.T) {
	decisionJSON := `{"explanation":"x","actions":[{"type":"run_command","summary":"s","danger":"safe","danger_reason":"dr","command":"echo hi","reason":"y"}]}`
	d, err := ParseDecision(decisionJSON)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.Actions[0].Type != ActionRunCommand {
		t.Fatalf("unexpected action type: %s", d.Actions[0].Type)
	}
}

func TestParseDecision_ValidReadOnlyInterrogation(t *testing.T) {
	decisionJSON := `{"explanation":"x","actions":[` +
		`{"type":"list_files","summary":"s1","danger":"safe","danger_reason":"dr1","path":""},` +
		`{"type":"read_files","summary":"s2","danger":"safe","danger_reason":"dr2","paths":["go.mod"]}` +
		`]}`
	if _, err := ParseDecision(decisionJSON); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseDecision_ReadFilesTooManyPathsRejected(t *testing.T) {
	decisionJSON := `{"explanation":"x","actions":[{"type":"read_files","summary":"s","danger":"safe","danger_reason":"dr","paths":["a","b","c","d","e","f","g"]}]}`
	if _, err := ParseDecision(decisionJSON); err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrchestrator_Apply_ApprovalGate(t *testing.T) {
	llm := &mockLLM{}
	approver := &mockApprover{ok: true}
	tools := &mockTools{}

	o, err := NewOrchestrator(llm, "model", approver, tools)
	if err != nil {
		t.Fatalf("new orchestrator: %v", err)
	}

	d := &Decision{
		Explanation: "x",
		Actions: []ProposedAction{
			{Type: ActionRunCommand, Summary: "s", Danger: DangerSafe, DangerReason: "dr", Command: "echo hi"},
			{Type: ActionEditFiles, Summary: "s2", Danger: DangerSafe, DangerReason: "dr2", Edits: []FileEdit{{Path: "a", Content: "b"}}},
		},
	}

	if err := o.Apply(context.Background(), d); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if tools.ran != 1 {
		t.Fatalf("expected 1 command run, got %d", tools.ran)
	}
	if tools.edits != 1 {
		t.Fatalf("expected 1 edit, got %d", tools.edits)
	}
}

func TestOrchestrator_Diagnose_ParsesDecision(t *testing.T) {
	llm := &mockLLM{text: `{"explanation":"x","actions":[{"type":"stop","summary":"s","danger":"safe","danger_reason":"dr","reason":"done"}]}`}
	approver := &mockApprover{ok: true}

	o, err := NewOrchestrator(llm, "model", approver, nil)
	if err != nil {
		t.Fatalf("new orchestrator: %v", err)
	}

	report := &setup.ExecutionReport{}
	d, err := o.Diagnose(context.Background(), report)
	if err != nil {
		t.Fatalf("diagnose: %v", err)
	}
	if d.Actions[0].Type != ActionStop {
		t.Fatalf("expected stop action")
	}
}

func TestOrchestrator_Diagnose_InterrogationToolErrorDoesNotAbort(t *testing.T) {
	llm := &seqLLM{texts: []string{
		`{"explanation":"x","actions":[{"type":"read_files","summary":"s","danger":"safe","danger_reason":"dr","paths":["/abs/outside"]}]}`,
		`{"explanation":"x2","actions":[{"type":"stop","summary":"s2","danger":"safe","danger_reason":"dr2","reason":"done"}]}`,
	}}
	approver := &mockApprover{ok: true}
	tools := &mockTools{}

	o, err := NewOrchestrator(llm, "model", approver, tools)
	if err != nil {
		t.Fatalf("new orchestrator: %v", err)
	}

	report := &setup.ExecutionReport{}
	if _, err := o.Diagnose(context.Background(), report); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
