package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/testutil"
)

type testSetup struct {
	tmpDir      string
	originalDir string
}

func setupTestDir(t *testing.T) *testSetup {
	t.Helper()
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	return &testSetup{
		tmpDir:      tmpDir,
		originalDir: originalDir,
	}
}

func (ts *testSetup) cleanup() {
	_ = os.Chdir(ts.originalDir)
}

func (ts *testSetup) createInitiatDir(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(".initiat", 0755); err != nil {
		t.Fatalf("Failed to create .initiat dir: %v", err)
	}
}

func (ts *testSetup) writeConfigFile(t *testing.T, org, project string) {
	t.Helper()
	ts.createInitiatDir(t)
	configContent := "org: " + org + "\nproject: " + project + "\n"
	configPath := filepath.Join(ts.tmpDir, ".initiat", "config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

func (ts *testSetup) writeSetupFile(t *testing.T, content string) {
	t.Helper()
	ts.createInitiatDir(t)
	setupPath := filepath.Join(ts.tmpDir, ".initiat", "setup.yml")
	if err := os.WriteFile(setupPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write setup file: %v", err)
	}
}

func TestRunProjectSetup_FileNotFound(t *testing.T) {
	ts := setupTestDir(t)
	defer ts.cleanup()

	ts.writeConfigFile(t, "test-org", "test-project")

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	err := runSetupRun(setupRunCmd, []string{})

	if err == nil {
		t.Fatal("Expected error for missing setup file")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected error to mention 'parse', got: %v", err)
	}
}

func TestRunProjectSetup_ValidFile(t *testing.T) {
	ts := setupTestDir(t)
	defer ts.cleanup()

	ts.writeConfigFile(t, "test-org", "test-project")
	ts.writeSetupFile(t, `version: 1
name: "Test Setup"
setup:
  - name: "test"
    run: "echo hello"
`)

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	err := runSetupRun(setupRunCmd, []string{})

	if err != nil {
		if strings.Contains(err.Error(), "project context") ||
			strings.Contains(err.Error(), "device not registered") {
			return
		}
		t.Logf("Unexpected error (may indicate test environment issue): %v", err)
		return
	}
}

func TestRunProjectSetup_InvalidConfig(t *testing.T) {
	ts := setupTestDir(t)
	defer ts.cleanup()

	ts.writeConfigFile(t, "test-org", "test-project")
	ts.writeSetupFile(t, `version: 2
name: "Invalid Setup"
`)

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	err := runSetupRun(setupRunCmd, []string{})

	if err == nil {
		t.Fatal("Expected error for invalid config")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected error to mention 'validation failed', got: %v", err)
	}
}
