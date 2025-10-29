package actions

import (
	"testing"
)

func TestPrintAction_Render(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		os           string
		expected     string
		expectedArgs []string
	}{
		{
			name:         "simple message on linux",
			message:      "Hello World",
			os:           "linux",
			expected:     "echo",
			expectedArgs: []string{"Hello World"},
		},
		{
			name:         "message with quotes on macos",
			message:      "Hello 'World'",
			os:           "macos",
			expected:     "echo",
			expectedArgs: []string{"Hello 'World'"},
		},
		{
			name:         "empty message on windows",
			message:      "",
			os:           "windows",
			expected:     "echo",
			expectedArgs: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewPrintAction(tt.message)
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
		})
	}
}

func TestPrintAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "valid message",
			message: "Hello World",
			wantErr: false,
		},
		{
			name:    "empty message",
			message: "",
			wantErr: false, // Print action should allow empty messages
		},
		{
			name:    "message with special characters",
			message: "Hello $USER & 'World'",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewPrintAction(tt.message)
			err := action.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintAction_Type(t *testing.T) {
	action := NewPrintAction("hello")
	if action.Type() != ActionTypePrint {
		t.Errorf("Expected type %v, got %v", ActionTypePrint, action.Type())
	}
}
