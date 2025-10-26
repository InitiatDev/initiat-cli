package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/env"
)

func TestEnvListCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := env.CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	err = env.CreateEnvironmentDir("dev")
	if err != nil {
		t.Fatalf("CreateEnvironmentDir failed: %v", err)
	}

	err = env.WriteSecrets("dev", "API_KEY=secret123")
	if err != nil {
		t.Fatalf("WriteSecrets failed: %v", err)
	}

	cmd := envListCmd
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("env list command failed: %v", err)
	}
}

func TestEnvSwitchCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := env.CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	err = env.CreateEnvironmentDir("dev")
	if err != nil {
		t.Fatalf("CreateEnvironmentDir failed: %v", err)
	}

	err = env.SetActiveEnvironment("dev")
	if err != nil {
		t.Fatalf("SetActiveEnvironment failed: %v", err)
	}

	activeEnv, err := env.GetActiveEnvironment()
	if err != nil {
		t.Fatalf("GetActiveEnvironment failed: %v", err)
	}

	if activeEnv != "dev" {
		t.Errorf("Expected active environment 'dev', got '%s'", activeEnv)
	}
}

func TestEnvCurrentCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := env.CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	err = env.CreateEnvironmentDir("dev")
	if err != nil {
		t.Fatalf("CreateEnvironmentDir failed: %v", err)
	}

	err = env.SetActiveEnvironment("dev")
	if err != nil {
		t.Fatalf("SetActiveEnvironment failed: %v", err)
	}

	cmd := envCurrentCmd
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("env current command failed: %v", err)
	}
}

func TestEnvUnsetCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := env.CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Create .git directory and configure gitignore to make init complete
	err = os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	err = env.EnsureGitignore()
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}

	err = env.CreateEnvironmentDir("dev")
	if err != nil {
		t.Fatalf("CreateEnvironmentDir failed: %v", err)
	}

	err = env.SetActiveEnvironment("dev")
	if err != nil {
		t.Fatalf("SetActiveEnvironment failed: %v", err)
	}

	// Verify environment is set
	activeEnv, err := env.GetActiveEnvironment()
	if err != nil {
		t.Fatalf("GetActiveEnvironment failed: %v", err)
	}
	if activeEnv != "dev" {
		t.Errorf("Expected active environment 'dev', got '%s'", activeEnv)
	}

	// Test the UnsetActiveEnvironment function directly
	err = env.UnsetActiveEnvironment()
	if err != nil {
		t.Fatalf("UnsetActiveEnvironment failed: %v", err)
	}

	// Verify environment is unset
	_, err = env.GetActiveEnvironment()
	if err == nil {
		t.Error("Expected GetActiveEnvironment to fail after unsetting")
	}
}

func TestEnvUnsetCommandNoActiveEnv(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := env.CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Create .git directory and configure gitignore to make init complete
	err = os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	err = env.EnsureGitignore()
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}

	// Test the UnsetActiveEnvironment function directly when no environment is set
	err = env.UnsetActiveEnvironment()
	if err != nil {
		t.Fatalf("UnsetActiveEnvironment should not fail when no environment is set: %v", err)
	}
}

func TestEnvInitCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Safety check: ensure we're in a temporary directory, not the project directory
	if !strings.Contains(tempDir, "/var/folders/") && !strings.Contains(tempDir, "/tmp/") {
		t.Fatalf("Test is not running in a temporary directory: %s", tempDir)
	}

	// Test the CreateInitiatDir function directly instead of the full command
	// This avoids the need for command execution setup
	err := env.CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	_, err = os.Stat(".initiat")
	if err != nil {
		t.Errorf("Expected .initiat directory to be created: %v", err)
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "just now"},
		{2 * time.Minute, "2m ago"},
		{2 * time.Hour, "2h ago"},
		{2 * 24 * time.Hour, "2d ago"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := formatTimeAgo(now.Add(-test.duration))
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}
