package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/testutil"
)

func TestRunSetupValidate_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupFile := filepath.Join(tmpDir, "setup.yml")

	yamlContent := `version: 1
name: "Test Setup"
setup:
  - name: "test"
    run: "echo hello"
`

	if err := os.WriteFile(setupFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	cmd := setupValidateCmd
	args := []string{setupFile}

	err := runSetupValidate(cmd, args)
	if err != nil {
		t.Errorf("runSetupValidate() error = %v", err)
	}

	capture.AssertContains(t, "Validating")
	capture.AssertContains(t, "Setup script is valid")
}

func TestRunSetupValidate_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupFile := filepath.Join(tmpDir, "setup.yml")

	yamlContent := `version: 2
name: "Invalid Setup"
`

	if err := os.WriteFile(setupFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	cmd := setupValidateCmd
	args := []string{setupFile}

	err := runSetupValidate(cmd, args)
	if err == nil {
		t.Error("runSetupValidate() expected error for invalid config")
	}

	capture.AssertContains(t, "Validation failed")
	capture.AssertContains(t, "version")
}

func TestRunSetupValidate_DefaultFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	if err := os.MkdirAll(".initiat", 0755); err != nil {
		t.Fatalf("Failed to create .initiat dir: %v", err)
	}

	setupFile := filepath.Join(tmpDir, ".initiat", "setup.yml")
	yamlContent := `version: 1`

	if err := os.WriteFile(setupFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	cmd := setupValidateCmd
	args := []string{}

	err := runSetupValidate(cmd, args)
	if err != nil {
		t.Errorf("runSetupValidate() error = %v", err)
	}

	capture.AssertContains(t, ".initiat/setup.yml")
	capture.AssertContains(t, "valid")
}

func TestRunSetupSchema_Stdout(t *testing.T) {
	t.Skip("Skipping stdout test due to blocking pipe issue - schema output is tested via file output")
}

func TestRunSetupSchema_ToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "schema.json")

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	schemaOutput = outputFile
	defer func() {
		schemaOutput = ""
	}()

	cmd := setupSchemaCmd
	args := []string{}

	err := runSetupSchema(cmd, args)
	if err != nil {
		t.Errorf("runSetupSchema() error = %v", err)
	}

	capture.AssertContains(t, "Schema saved to")

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Expected schema file to be created at %s", outputFile)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	contentStr := string(content)
	if !containsString(contentStr, "$schema") {
		t.Error("Expected schema file to contain $schema")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > 0 && len(substr) > 0 && indexOfString(s, substr) >= 0))
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
