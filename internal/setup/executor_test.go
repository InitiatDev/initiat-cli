package setup

import (
	"errors"
	"testing"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/testutil"
)

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
				"Executing setup script",
				"setup",
				"test",
				"completed successfully",
			},
		},
		{
			name: "multiple commands",
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
				"bootstrap",
				"setup",
				"step1",
				"step2",
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
				"completed successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := testutil.CaptureStdout()
			defer capture.Restore()

			mockExecutor := newMockCommandExecutor()
			executor := NewExecutorWithCommandExecutor(tt.secrets, mockExecutor)
			err := executor.Execute(tt.plan)

			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}

			for _, text := range tt.contains {
				capture.AssertContains(t, text)
			}

			for _, text := range tt.notContains {
				capture.AssertNotContains(t, text)
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

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	mockExecutor := newMockCommandExecutor()
	mockExecutor.SetError("false", errors.New("command failed"))
	executor := NewExecutorWithCommandExecutor(nil, mockExecutor)
	err := executor.Execute(plan)

	if err != nil {
		t.Errorf("Execute() should not fail with continue_on_error, got error: %v", err)
	}

	capture.AssertContains(t, "success")
	capture.AssertContains(t, "Command failed but continuing")
	capture.AssertContains(t, "final")
	capture.AssertContains(t, "completed successfully")
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

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	mockExecutor := newMockCommandExecutor()
	mockExecutor.SetError("false", errors.New("command failed"))
	executor := NewExecutorWithCommandExecutor(nil, mockExecutor)
	err := executor.Execute(plan)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	capture.AssertContains(t, "completed successfully")
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
	mockExecutor := newMockCommandExecutor()
	executor := NewExecutorWithCommandExecutor(nil, mockExecutor)

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

	executed := mockExecutor.GetExecuted()
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

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOfString(s, substr) >= 0)))
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
