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
