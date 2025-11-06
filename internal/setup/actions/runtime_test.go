package actions

import (
	"strings"
	"testing"
	"time"
)

func TestEnsureRuntimeAction_Validate(t *testing.T) {
	tests := []struct {
		name        string
		runtimeName string
		version     string
		manager     *RuntimeManager
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid node with asdf",
			runtimeName: "node",
			version:     "18.0.0",
			manager:     &RuntimeManager{Asdf: true},
			expectError: false,
		},
		{
			name:        "empty runtime name",
			runtimeName: "",
			expectError: true,
			errorMsg:    "runtime name cannot be empty",
		},
		{
			name:        "invalid runtime name",
			runtimeName: "invalid-runtime",
			expectError: true,
			errorMsg:    "invalid runtime name",
		},
		{
			name:        "no install method",
			runtimeName: "node",
			expectError: true,
			errorMsg:    "at least one install method must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewEnsureRuntimeAction(tt.runtimeName, tt.version, tt.manager)
			err := action.Validate()

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestEnsureRuntimeAction_Render(t *testing.T) {
	tests := []struct {
		name        string
		runtimeName string
		version     string
		manager     *RuntimeManager
		os          string
		expectError bool
	}{
		{
			name:        "node with asdf on macOS",
			runtimeName: "node",
			version:     "18.0.0",
			manager:     &RuntimeManager{Asdf: true},
			os:          OSMacOS,
			expectError: false,
		},
		{
			name:        "empty runtime name",
			runtimeName: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewEnsureRuntimeAction(tt.runtimeName, tt.version, tt.manager)
			ctx := &ActionContext{
				OS:         tt.os,
				Arch:       "x86_64",
				Env:        map[string]string{},
				WorkingDir: "/tmp",
				Timeout:    30 * time.Second,
			}

			commands, err := action.Render(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(commands) == 0 {
					t.Error("Expected commands but got none")
				}
			}
		})
	}
}

// TestEnsureRuntimeAction_GetBrewFormula removed - logic now handled by strategy pattern

// TestEnsureRuntimeAction_GetAptPackage removed - logic now handled by strategy pattern

// TestEnsureRuntimeAction_GetChocoPackage removed - logic now handled by strategy pattern

func TestEnsureRuntimeAction_StrategyBasedCommands(t *testing.T) {
	manager := &RuntimeManager{Asdf: true}
	action := NewEnsureRuntimeAction("node", "18.0.0", manager)
	ctx := &ActionContext{
		OS:         OSMacOS,
		Arch:       "x86_64",
		Env:        map[string]string{},
		WorkingDir: "/tmp",
		Timeout:    30 * time.Second,
	}

	commands, err := action.getInstallCommands(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Should have plugin add, install, and global commands for asdf
	expectedCommands := []string{"asdf", "asdf", "asdf"}
	if len(commands) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(commands))
	}

	for i, expected := range expectedCommands {
		if commands[i].Command != expected {
			t.Errorf("Command %d: expected %s, got %s", i, expected, commands[i].Command)
		}
	}
}

// Old test methods removed - functionality now handled by strategy pattern
