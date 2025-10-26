package env

import (
	"os"
	"strings"
	"testing"
)

func TestEnsureGitignore(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := EnsureGitignore()
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}

	content, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	if !strings.Contains(string(content), "# Initiat") {
		t.Error("Expected Initiat entry in .gitignore")
	}

	if !strings.Contains(string(content), ".initiat/environments/*/secrets.env") {
		t.Error("Expected secrets.env pattern in .gitignore")
	}

	if !strings.Contains(string(content), ".initiat/active") {
		t.Error("Expected active file pattern in .gitignore")
	}
}

func TestEnsureGitignoreIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := EnsureGitignore()
	if err != nil {
		t.Fatalf("First EnsureGitignore failed: %v", err)
	}

	firstContent, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore after first run: %v", err)
	}

	err = EnsureGitignore()
	if err != nil {
		t.Fatalf("Second EnsureGitignore failed: %v", err)
	}

	secondContent, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore after second run: %v", err)
	}

	if string(firstContent) != string(secondContent) {
		t.Error("EnsureGitignore should be idempotent")
	}
}

func TestCheckGitignore(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	hasEntry, err := CheckGitignore()
	if err != nil {
		t.Fatalf("CheckGitignore failed: %v", err)
	}

	if hasEntry {
		t.Error("Expected no gitignore entry initially")
	}

	err = EnsureGitignore()
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}

	hasEntry, err = CheckGitignore()
	if err != nil {
		t.Fatalf("CheckGitignore failed after adding entry: %v", err)
	}

	if !hasEntry {
		t.Error("Expected gitignore entry after EnsureGitignore")
	}
}

func TestIsGitRepository(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	if IsGitRepository() {
		t.Error("Expected not a git repository initially")
	}

	err := os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	if !IsGitRepository() {
		t.Error("Expected git repository after creating .git")
	}
}

func TestGetGitignoreStatus(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	status, err := GetGitignoreStatus()
	if err != nil {
		t.Fatalf("GetGitignoreStatus failed: %v", err)
	}

	if status != "not a git repository" {
		t.Errorf("Expected 'not a git repository', got '%s'", status)
	}

	err = os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	status, err = GetGitignoreStatus()
	if err != nil {
		t.Fatalf("GetGitignoreStatus failed after creating .git: %v", err)
	}

	if status != "missing" {
		t.Errorf("Expected 'missing', got '%s'", status)
	}

	err = EnsureGitignore()
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}

	status, err = GetGitignoreStatus()
	if err != nil {
		t.Fatalf("GetGitignoreStatus failed after EnsureGitignore: %v", err)
	}

	if status != "configured" {
		t.Errorf("Expected 'configured', got '%s'", status)
	}
}
