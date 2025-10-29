package actions

import (
	"testing"
)

func TestAssertCommandAction_Render(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		os           string
		expected     string
		expectedArgs []string
	}{
		{
			name:         "which command on linux",
			command:      "which git",
			os:           "linux",
			expected:     "which",
			expectedArgs: []string{"git"},
		},
		{
			name:         "command with args on macos",
			command:      "ls -la /tmp",
			os:           "macos",
			expected:     "ls",
			expectedArgs: []string{"-la", "/tmp"},
		},
		{
			name:         "where command on windows",
			command:      "where git",
			os:           "windows",
			expected:     "where",
			expectedArgs: []string{"git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewAssertCommandAction(tt.command)
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

func TestAssertCommandAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "which git",
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
			action := NewAssertCommandAction(tt.command)
			err := action.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertCommandAction_Type(t *testing.T) {
	action := NewAssertCommandAction("which git")
	if action.Type() != ActionTypeAssertCommand {
		t.Errorf("Expected type %v, got %v", ActionTypeAssertCommand, action.Type())
	}
}
