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
		fallbacks   []RuntimeFallbackInstaller
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
			name:        "valid python with brew fallback",
			runtimeName: "python",
			version:     "3.11",
			fallbacks: []RuntimeFallbackInstaller{
				{Brew: &BrewInstall{Formula: "python"}},
			},
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
			action := NewEnsureRuntimeAction(tt.runtimeName, tt.version, tt.manager, tt.fallbacks)
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
		fallbacks   []RuntimeFallbackInstaller
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
			name:        "python with brew on macOS",
			runtimeName: "python",
			version:     "3.11",
			fallbacks: []RuntimeFallbackInstaller{
				{Brew: &BrewInstall{Formula: "python"}},
			},
			os:          OSMacOS,
			expectError: false,
		},
		{
			name:        "go with apt on Linux",
			runtimeName: "go",
			version:     "1.21",
			fallbacks: []RuntimeFallbackInstaller{
				{Apt: &AptInstall{Packages: []string{"golang-go"}}},
			},
			os:          OSLinux,
			expectError: false,
		},
		{
			name:        "rust with choco on Windows",
			runtimeName: "rust",
			version:     "1.70",
			fallbacks: []RuntimeFallbackInstaller{
				{Choco: &ChocoInstall{Packages: []string{"rust"}}},
			},
			os:          OSWindows,
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
			action := NewEnsureRuntimeAction(tt.runtimeName, tt.version, tt.manager, tt.fallbacks)
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

func TestEnsureRuntimeAction_GetBrewFormula(t *testing.T) {
	tests := []struct {
		runtimeName string
		expected    string
	}{
		{"node", "node"},
		{"python", "python"},
		{"go", "go"},
		{"elixir", "elixir"},
		{"erlang", "erlang"},
		{"java", "openjdk"},
		{"rust", "rust"},
		{"dotnet", "dotnet"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.runtimeName, func(t *testing.T) {
			action := &EnsureRuntimeAction{runtimeName: tt.runtimeName}
			result := action.getBrewFormula()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnsureRuntimeAction_GetAptPackage(t *testing.T) {
	tests := []struct {
		runtimeName string
		expected    string
	}{
		{"node", "nodejs"},
		{"python", "python3"},
		{"go", "golang-go"},
		{"elixir", "elixir"},
		{"erlang", "erlang"},
		{"java", "openjdk-11-jdk"},
		{"rust", "rustc"},
		{"dotnet", "dotnet-sdk-6.0"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.runtimeName, func(t *testing.T) {
			action := &EnsureRuntimeAction{runtimeName: tt.runtimeName}
			result := action.getAptPackage()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnsureRuntimeAction_GetChocoPackage(t *testing.T) {
	tests := []struct {
		runtimeName string
		expected    string
	}{
		{"node", "nodejs"},
		{"python", "python"},
		{"go", "golang"},
		{"elixir", "elixir"},
		{"erlang", "erlang"},
		{"java", "openjdk"},
		{"rust", "rust"},
		{"dotnet", "dotnet"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.runtimeName, func(t *testing.T) {
			action := &EnsureRuntimeAction{runtimeName: tt.runtimeName}
			result := action.getChocoPackage()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnsureRuntimeAction_AsdfCommands(t *testing.T) {
	action := &EnsureRuntimeAction{
		runtimeName: "node",
		version:     "18.0.0",
	}
	ctx := &ActionContext{
		OS:   OSMacOS,
		Arch: "x86_64",
	}

	commands, err := action.getAsdfCommands(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"list", "plugin add", "install", "global"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd.Command == "asdf" && strings.Contains(strings.Join(cmd.Args, " "), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected asdf command containing '%s' not found", expected)
		}
	}
}

func TestEnsureRuntimeAction_BrewCommands(t *testing.T) {
	action := &EnsureRuntimeAction{
		runtimeName: "node",
		version:     "18.0.0",
	}
	ctx := &ActionContext{
		OS:   OSMacOS,
		Arch: "x86_64",
	}
	brew := &BrewInstall{Formula: "node"}

	commands, err := action.getBrewCommands(brew, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"which", "brew install"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if strings.Contains(cmd.Command+" "+strings.Join(cmd.Args, " "), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command containing '%s' not found", expected)
		}
	}
}

func TestEnsureRuntimeAction_AptCommands(t *testing.T) {
	action := &EnsureRuntimeAction{
		runtimeName: "node",
		version:     "18.0.0",
	}
	ctx := &ActionContext{
		OS:   OSLinux,
		Arch: "x86_64",
	}
	apt := &AptInstall{Packages: []string{"nodejs"}, Update: true}

	commands, err := action.getAptCommands(apt, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"which", "apt update", "apt install"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if strings.Contains(cmd.Command+" "+strings.Join(cmd.Args, " "), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command containing '%s' not found", expected)
		}
	}
}

func TestEnsureRuntimeAction_ChocoCommands(t *testing.T) {
	action := &EnsureRuntimeAction{
		runtimeName: "node",
		version:     "18.0.0",
	}
	ctx := &ActionContext{
		OS:   OSWindows,
		Arch: "x86_64",
	}
	choco := &ChocoInstall{Packages: []string{"nodejs"}}

	commands, err := action.getChocoCommands(choco, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"where", "choco install"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if strings.Contains(cmd.Command+" "+strings.Join(cmd.Args, " "), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command containing '%s' not found", expected)
		}
	}
}
