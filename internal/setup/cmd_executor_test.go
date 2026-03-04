package setup

import (
	"context"
	"testing"
	"time"
)

type mockCommandExecutor struct {
	executed []*CommandRequest
	errors   map[string]error
	results  map[string]*CommandResult
}

func newMockCommandExecutor() *mockCommandExecutor {
	return &mockCommandExecutor{
		executed: []*CommandRequest{},
		errors:   make(map[string]error),
		results:  make(map[string]*CommandResult),
	}
}

func (m *mockCommandExecutor) SetError(command string, err error) {
	m.errors[command] = err
}

func (m *mockCommandExecutor) SetResult(command string, res *CommandResult) {
	m.results[command] = res
}

func (m *mockCommandExecutor) Execute(ctx context.Context, req *CommandRequest) (*CommandResult, error) {
	m.executed = append(m.executed, req)

	res := m.results[req.Command]
	if res == nil {
		res = &CommandResult{ExitCode: 0}
	}

	if err, ok := m.errors[req.Command]; ok {
		if res.ExitCode == 0 {
			res.ExitCode = 1
		}
		return res, err
	}

	return res, nil
}

func (m *mockCommandExecutor) GetExecuted() []*CommandRequest {
	return m.executed
}

func TestRealCommandExecutor_Execute(t *testing.T) {
	executor := NewRealCommandExecutor()

	req := &CommandRequest{
		Command:    "echo",
		Args:       []string{"test"},
		Env:        map[string]string{"TEST_VAR": "value"},
		WorkingDir: "",
		Timeout:    0,
	}

	ctx := context.Background()
	res, err := executor.Execute(ctx, req)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if res == nil {
		t.Fatalf("expected result, got nil")
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
}

func TestRealCommandExecutor_Execute_Timeout(t *testing.T) {
	executor := NewRealCommandExecutor()

	req := &CommandRequest{
		Command:    "sleep",
		Args:       []string{"2"},
		Timeout:    time.Millisecond * 100,
		WorkingDir: "",
	}

	ctx := context.Background()
	res, err := executor.Execute(ctx, req)
	if err == nil {
		t.Error("Execute() expected timeout error, got nil")
	}
	if res == nil {
		t.Fatalf("expected result, got nil")
	}
	if !res.TimedOut {
		t.Fatalf("expected TimedOut=true")
	}
}

func TestBuildCommandLine(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		args     []string
		expected string
	}{
		{
			name:     "command without args",
			command:  "echo",
			args:     []string{},
			expected: "echo",
		},
		{
			name:     "command with single arg",
			command:  "echo",
			args:     []string{"hello"},
			expected: "echo hello",
		},
		{
			name:     "command with multiple args",
			command:  "ls",
			args:     []string{"-la", "/tmp"},
			expected: "ls -la /tmp",
		},
		{
			name:     "command with space in arg",
			command:  "echo",
			args:     []string{"hello world"},
			expected: `echo "hello world"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCommandLine(tt.command, tt.args)
			if got != tt.expected {
				t.Errorf("buildCommandLine(%q, %v) = %q, want %q", tt.command, tt.args, got, tt.expected)
			}
		})
	}
}
