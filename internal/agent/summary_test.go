package agent

import (
	"context"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type stubLLM struct {
	text string
	err  error
}

func (s *stubLLM) Complete(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &CompleteResponse{Text: s.text}, nil
}

func TestAssessIssues_ParsesBuckets(t *testing.T) {
	llm := &stubLLM{
		text: `{"setup_or_app":["missing go mod tidy"],"local_environment":["brew not installed"],"notes":"x"}`,
	}

	report := &setup.ExecutionReport{}
	decision := &Decision{Explanation: "x", Actions: []ProposedAction{{Type: ActionStop}}}

	got, err := AssessIssues(context.Background(), llm, "model", report, decision)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got.SetupOrApp) != 1 || got.SetupOrApp[0] != "missing go mod tidy" {
		t.Fatalf("unexpected setup_or_app: %#v", got.SetupOrApp)
	}
	if len(got.LocalEnvironment) != 1 || got.LocalEnvironment[0] != "brew not installed" {
		t.Fatalf("unexpected local_environment: %#v", got.LocalEnvironment)
	}
	if got.Notes != "x" {
		t.Fatalf("unexpected notes: %q", got.Notes)
	}
}
