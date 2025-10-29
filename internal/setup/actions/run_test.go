package actions

import (
	"testing"
)

func TestRunAction_Render(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		os           string
		expected     string
		expectedArgs []string
	}{
		{
			name:         "simple command on linux",
			command:      "echo hello",
			os:           "linux",
			expected:     "echo",
			expectedArgs: []string{"hello"},
		},
		{
			name:         "complex command on macos",
			command:      "ls -la /tmp",
			os:           "macos",
			expected:     "ls",
			expectedArgs: []string{"-la", "/tmp"},
		},
		{
			name:         "command with spaces on windows",
			command:      "dir C:\\Users",
			os:           "windows",
			expected:     "dir",
			expectedArgs: []string{"C:\\Users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewRunAction(tt.command)
			ctx := &ActionContext{OS: tt.os, Arch: "amd64"}

			commands, err := action.Render(ctx)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			if len(commands) != 1 {
				t.Fatalf("Expected 1 command, got %d", len(commands))
			}

			if commands[0].Command != tt.expected {
				t.Errorf("Expected command %q, got %q", tt.expected, commands[0].Command)
			}

			if len(commands[0].Args) != len(tt.expectedArgs) {
				t.Fatalf("Expected %d args, got %d", len(tt.expectedArgs), len(commands[0].Args))
			}

			for i, expectedArg := range tt.expectedArgs {
				if commands[0].Args[i] != expectedArg {
					t.Errorf("Expected arg[%d] %q, got %q", i, expectedArg, commands[0].Args[i])
				}
			}

			if commands[0].Description == "" {
				t.Error("Expected non-empty description")
			}
		})
	}
}

func TestRunAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:    "whitespace only command",
			command: "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewRunAction(tt.command)
			err := action.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAction_Type(t *testing.T) {
	action := NewRunAction("echo hello")
	if action.Type() != ActionTypeRun {
		t.Errorf("Expected type %v, got %v", ActionTypeRun, action.Type())
	}
}
