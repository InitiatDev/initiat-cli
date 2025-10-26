package env

import (
	"os"
	"testing"
)

func TestIsInitCompleted(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Initially should not be completed
	if IsInitCompleted() {
		t.Error("Expected init not to be completed initially")
	}

	// Create .initiat directory
	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Should still not be completed (no gitignore)
	if IsInitCompleted() {
		t.Error("Expected init not to be completed without gitignore")
	}

	// Create .git directory to make it a git repo
	err = os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Should still not be completed (no gitignore configured)
	if IsInitCompleted() {
		t.Error("Expected init not to be completed without configured gitignore")
	}

	// Configure gitignore
	err = EnsureGitignore()
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}

	// Now should be completed
	if !IsInitCompleted() {
		t.Error("Expected init to be completed after setting up gitignore")
	}
}

func TestIsInitCompletedWithoutGit(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create .initiat directory
	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Should not be completed without git repository
	if IsInitCompleted() {
		t.Error("Expected init not to be completed without git repository")
	}
}

func TestIsInitCompletedWithGitButNoGitignore(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create .initiat directory
	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Create .git directory to make it a git repo
	err = os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Should not be completed without configured gitignore
	if IsInitCompleted() {
		t.Error("Expected init not to be completed without configured gitignore")
	}
}

func TestUnsetActiveEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create .initiat directory
	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Create environments directory
	err = CreateEnvironmentDir("test-env")
	if err != nil {
		t.Fatalf("CreateEnvironmentDir failed: %v", err)
	}

	// Set an active environment first
	err = SetActiveEnvironment("test-env")
	if err != nil {
		t.Fatalf("SetActiveEnvironment failed: %v", err)
	}

	// Verify it's set
	activeEnv, err := GetActiveEnvironment()
	if err != nil {
		t.Fatalf("GetActiveEnvironment failed: %v", err)
	}
	if activeEnv != "test-env" {
		t.Errorf("Expected active environment to be 'test-env', got '%s'", activeEnv)
	}

	// Unset the active environment
	err = UnsetActiveEnvironment()
	if err != nil {
		t.Fatalf("UnsetActiveEnvironment failed: %v", err)
	}

	// Verify it's unset
	_, err = GetActiveEnvironment()
	if err == nil {
		t.Error("Expected GetActiveEnvironment to fail after unsetting")
	}
}

func TestUnsetActiveEnvironmentWhenNoneSet(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create .initiat directory
	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	// Try to unset when no environment is set
	err = UnsetActiveEnvironment()
	if err != nil {
		t.Fatalf("UnsetActiveEnvironment should not fail when no environment is set: %v", err)
	}
}
