package setup

import (
	"context"
	"testing"
	"time"
)

type mockCommandExecutor struct {
	executed []*CommandRequest
	errors   map[string]error
}

func newMockCommandExecutor() *mockCommandExecutor {
	return &mockCommandExecutor{
		executed: []*CommandRequest{},
		errors:   make(map[string]error),
	}
}

func (m *mockCommandExecutor) SetError(command string, err error) {
	m.errors[command] = err
}

func (m *mockCommandExecutor) Execute(ctx context.Context, req *CommandRequest) error {
	m.executed = append(m.executed, req)

	if err, ok := m.errors[req.Command]; ok {
		return err
	}

	return nil
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
	err := executor.Execute(ctx, req)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
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
	err := executor.Execute(ctx, req)
	if err == nil {
		t.Error("Execute() expected timeout error, got nil")
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
