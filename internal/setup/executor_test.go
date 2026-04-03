package setup

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/output"
)

func newTestExecutor(secrets map[string]string, mock *mockCommandExecutor) (*Executor, *bytes.Buffer) {
	var buf bytes.Buffer
	f := output.NewFormatter(&buf, output.WithColor(false), output.WithFancy(false))
	return NewExecutorWithFormatter(secrets, mock, f), &buf
}

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name        string
		plan        *ExecutionPlan
		secrets     map[string]string
		wantError   bool
		contains    []string
		notContains []string
	}{
		{
			name: "single command execution",
			plan: &ExecutionPlan{
				Commands: []ExecutableCommand{
					{
						Phase:           "setup",
						StepName:        "test",
						Command:         "echo",
						Args:            []string{"hello"},
						Description:     "Test command",
						ContinueOnError: false,
					},
				},
				Summary: ExecutionSummary{
					TotalSteps:    1,
					TotalCommands: 1,
					Phases: []PhaseSummary{
						{Name: "setup", StepCount: 1, CommandCount: 1},
					},
				},
			},
			secrets:   nil,
			wantError: false,
			contains: []string{
				"== setup ==",
				"[ok] test",
			},
		},
		{
			name: "multiple commands across phases",
			plan: &ExecutionPlan{
				Commands: []ExecutableCommand{
					{
						Phase:           "bootstrap",
						StepName:        "step1",
						Command:         "echo",
						Args:            []string{"step1"},
						ContinueOnError: false,
					},
					{
						Phase:           "setup",
						StepName:        "step2",
						Command:         "echo",
						Args:            []string{"step2"},
						ContinueOnError: false,
					},
				},
				Summary: ExecutionSummary{
					TotalSteps:    2,
					TotalCommands: 2,
					Phases: []PhaseSummary{
						{Name: "bootstrap", StepCount: 1, CommandCount: 1},
						{Name: "setup", StepCount: 1, CommandCount: 1},
					},
				},
			},
			secrets:   nil,
			wantError: false,
			contains: []string{
				"== bootstrap ==",
				"[ok] step1",
				"== setup ==",
				"[ok] step2",
			},
		},
		{
			name: "command with secrets redaction",
			plan: &ExecutionPlan{
				Commands: []ExecutableCommand{
					{
						Phase:           "setup",
						StepName:        "test",
						Command:         "echo",
						Args:            []string{"test"},
						Env:             map[string]string{"API_KEY": "secret123"},
						ContinueOnError: false,
					},
				},
				Summary: ExecutionSummary{
					TotalSteps:    1,
					TotalCommands: 1,
				},
			},
			secrets:   map[string]string{"API_KEY": "secret123"},
			wantError: false,
			contains: []string{
				"[ok] test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockCommandExecutor()
			executor, buf := newTestExecutor(tt.secrets, mock)
			err := executor.Execute(tt.plan)

			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}

			out := buf.String()
			for _, text := range tt.contains {
				if !strings.Contains(out, text) {
					t.Errorf("Expected output to contain %q, got:\n%s", text, out)
				}
			}

			for _, text := range tt.notContains {
				if strings.Contains(out, text) {
					t.Errorf("Expected output NOT to contain %q, got:\n%s", text, out)
				}
			}
		})
	}
}

func TestExecutor_Execute_ContinueOnError(t *testing.T) {
	plan := &ExecutionPlan{
		Commands: []ExecutableCommand{
			{
				Phase:           "setup",
				StepName:        "success",
				Command:         "echo",
				Args:            []string{"success"},
				ContinueOnError: false,
			},
			{
				Phase:           "setup",
				StepName:        "fail",
				Command:         "false",
				Args:            []string{},
				ContinueOnError: true,
			},
			{
				Phase:           "setup",
				StepName:        "final",
				Command:         "echo",
				Args:            []string{"final"},
				ContinueOnError: false,
			},
		},
		Summary: ExecutionSummary{
			TotalSteps:    3,
			TotalCommands: 3,
		},
	}

	mock := newMockCommandExecutor()
	mock.SetError("false", errors.New("command failed"))
	executor, buf := newTestExecutor(nil, mock)
	err := executor.Execute(plan)

	if err != nil {
		t.Errorf("Execute() should not fail with continue_on_error, got error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[ok] success") {
		t.Errorf("Expected success step, got:\n%s", out)
	}
	if !strings.Contains(out, "[FAIL] fail") {
		t.Errorf("Expected failure step, got:\n%s", out)
	}
	if !strings.Contains(out, "[ok] final") {
		t.Errorf("Expected final step, got:\n%s", out)
	}
}

func TestExecutor_Execute_Retry(t *testing.T) {
	plan := &ExecutionPlan{
		Commands: []ExecutableCommand{
			{
				Phase:           "setup",
				StepName:        "test",
				Command:         "echo",
				Args:            []string{"test"},
				Retries:         &RetryPolicy{Attempts: 3, Backoff: time.Millisecond * 10},
				ContinueOnError: false,
			},
		},
		Summary: ExecutionSummary{
			TotalSteps:    1,
			TotalCommands: 1,
		},
	}

	mock := newMockCommandExecutor()
	executor, buf := newTestExecutor(nil, mock)
	err := executor.Execute(plan)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[ok] test") {
		t.Errorf("Expected success output, got:\n%s", out)
	}
}

func TestExecutor_Execute_FailureShowsStderr(t *testing.T) {
	plan := &ExecutionPlan{
		Commands: []ExecutableCommand{
			{
				Phase:    "provision",
				StepName: "Run migrations",
				Command:  "migrate",
				Args:     []string{"up"},
			},
		},
		Summary: ExecutionSummary{
			TotalSteps:    1,
			TotalCommands: 1,
			Phases: []PhaseSummary{
				{Name: "provision", StepCount: 1, CommandCount: 1},
				{Name: "setup", StepCount: 1, CommandCount: 1},
			},
		},
	}

	mock := newMockCommandExecutor()
	mock.SetError("migrate", errors.New("migration failed"))
	mock.SetResult("migrate", &CommandResult{
		ExitCode: 1,
		Stderr:   "relation \"users\" already exists",
	})
	executor, buf := newTestExecutor(nil, mock)
	err := executor.Execute(plan)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	out := buf.String()
	if !strings.Contains(out, "[FAIL] Run migrations") {
		t.Errorf("Expected failure step, got:\n%s", out)
	}
	if !strings.Contains(out, "relation \"users\" already exists") {
		t.Errorf("Expected stderr in output, got:\n%s", out)
	}
}

func TestExecutor_Execute_SkippedPhases(t *testing.T) {
	plan := &ExecutionPlan{
		Commands: []ExecutableCommand{
			{
				Phase:    "bootstrap",
				StepName: "install",
				Command:  "fail-cmd",
			},
		},
		Summary: ExecutionSummary{
			TotalSteps:    3,
			TotalCommands: 3,
			Phases: []PhaseSummary{
				{Name: "bootstrap", StepCount: 1, CommandCount: 1},
				{Name: "provision", StepCount: 1, CommandCount: 1},
				{Name: "setup", StepCount: 1, CommandCount: 1},
			},
		},
	}

	mock := newMockCommandExecutor()
	mock.SetError("fail-cmd", errors.New("failed"))
	executor, buf := newTestExecutor(nil, mock)
	_ = executor.Execute(plan)

	out := buf.String()
	if !strings.Contains(out, "provision") || !strings.Contains(out, "setup") || !strings.Contains(out, "skipped") {
		t.Errorf("Expected skipped phases in output, got:\n%s", out)
	}
}

func TestExecutor_shouldRedact(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		secrets map[string]string
		want    bool
	}{
		{
			name:    "redact by secret map",
			key:     "API_KEY",
			secrets: map[string]string{"API_KEY": "secret"},
			want:    true,
		},
		{
			name:    "redact by keyword",
			key:     "PASSWORD",
			secrets: nil,
			want:    true,
		},
		{
			name:    "redact by keyword token",
			key:     "ACCESS_TOKEN",
			secrets: nil,
			want:    true,
		},
		{
			name:    "do not redact normal env var",
			key:     "NORMAL_VAR",
			secrets: nil,
			want:    false,
		},
		{
			name:    "redact secret in key name",
			key:     "MY_SECRET_VALUE",
			secrets: nil,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.secrets)
			if got := executor.shouldRedact(tt.key); got != tt.want {
				t.Errorf("shouldRedact(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestExecutor_CommandExecution(t *testing.T) {
	mock := newMockCommandExecutor()
	executor, _ := newTestExecutor(nil, mock)

	plan := &ExecutionPlan{
		Commands: []ExecutableCommand{
			{
				Phase:           "setup",
				StepName:        "test",
				Command:         "echo",
				Args:            []string{"hello"},
				Env:             map[string]string{"VAR": "value"},
				WorkingDir:      "/tmp",
				ContinueOnError: false,
			},
		},
		Summary: ExecutionSummary{
			TotalSteps:    1,
			TotalCommands: 1,
		},
	}

	err := executor.Execute(plan)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	executed := mock.GetExecuted()
	if len(executed) != 1 {
		t.Errorf("Expected 1 command execution, got %d", len(executed))
	}

	req := executed[0]
	if req.Command != "echo" {
		t.Errorf("Expected command 'echo', got %q", req.Command)
	}
	if len(req.Args) != 1 || req.Args[0] != "hello" {
		t.Errorf("Expected args ['hello'], got %v", req.Args)
	}
	if req.Env["VAR"] != "value" {
		t.Errorf("Expected env VAR=value, got %v", req.Env)
	}
	if req.WorkingDir != "/tmp" {
		t.Errorf("Expected working dir /tmp, got %q", req.WorkingDir)
	}
}

func TestExecutor_Execute_PhaseTransitions(t *testing.T) {
	plan := &ExecutionPlan{
		Commands: []ExecutableCommand{
			{Phase: "bootstrap", StepName: "install deps", Command: "echo"},
			{Phase: "bootstrap", StepName: "check versions", Command: "echo"},
			{Phase: "setup", StepName: "build", Command: "echo"},
		},
		Summary: ExecutionSummary{
			TotalSteps:    3,
			TotalCommands: 3,
			Phases: []PhaseSummary{
				{Name: "bootstrap", StepCount: 2, CommandCount: 2},
				{Name: "setup", StepCount: 1, CommandCount: 1},
			},
		},
	}

	mock := newMockCommandExecutor()
	executor, buf := newTestExecutor(nil, mock)
	err := executor.Execute(plan)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	// Verify phase ordering: bootstrap appears before setup
	bootstrapIdx := strings.Index(out, "== bootstrap ==")
	setupIdx := strings.Index(out, "== setup ==")
	if bootstrapIdx < 0 || setupIdx < 0 {
		t.Fatalf("Expected both phase headers, got:\n%s", out)
	}
	if bootstrapIdx >= setupIdx {
		t.Errorf("Expected bootstrap before setup, got:\n%s", out)
	}

	// Both steps in bootstrap phase appear between bootstrap and setup headers
	installIdx := strings.Index(out, "install deps")
	checkIdx := strings.Index(out, "check versions")
	if installIdx < bootstrapIdx || installIdx > setupIdx {
		t.Errorf("Expected 'install deps' within bootstrap phase, got:\n%s", out)
	}
	if checkIdx < bootstrapIdx || checkIdx > setupIdx {
		t.Errorf("Expected 'check versions' within bootstrap phase, got:\n%s", out)
	}
}
