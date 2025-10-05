package export

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitHandler_FindGitRoot(t *testing.T) {
	handler := NewGitHandler()

	tempDir := t.TempDir()

	gitRoot := filepath.Join(tempDir, "project")
	nestedPath := filepath.Join(gitRoot, "deep", "nested", "file.txt")

	os.MkdirAll(filepath.Join(gitRoot, ".git"), 0755)
	os.MkdirAll(filepath.Dir(nestedPath), 0755)

	found, exists := handler.FindGitRoot(nestedPath)
	if !exists {
		t.Error("Expected to find git root")
	}
	if found != gitRoot {
		t.Errorf("Expected git root %q, got %q", gitRoot, found)
	}
}

func TestGitHandler_FindGitRoot_NotFound(t *testing.T) {
	handler := NewGitHandler()

	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "no-git", "file.txt")
	os.MkdirAll(filepath.Dir(testPath), 0755)

	_, exists := handler.FindGitRoot(testPath)
	if exists {
		t.Error("Expected not to find git root")
	}
}

func TestGitHandler_ReadGitignore(t *testing.T) {
	handler := NewGitHandler()

	tempDir := t.TempDir()
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	content := "*.log\nsecrets/\n.env"

	os.WriteFile(gitignorePath, []byte(content), 0644)

	result, err := handler.ReadGitignore(tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != content {
		t.Errorf("Expected %q, got %q", content, result)
	}
}

func TestGitHandler_ReadGitignore_NotExists(t *testing.T) {
	handler := NewGitHandler()

	tempDir := t.TempDir()

	result, err := handler.ReadGitignore(tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestGitHandler_IsFileIgnored(t *testing.T) {
	handler := NewGitHandler()

	gitignoreContent := "*.log\nsecrets/\n.env\nconfig/secrets.txt"

	tempDir := t.TempDir()
	ignoredFile := filepath.Join(tempDir, "secrets", "file.txt")
	notIgnoredFile := filepath.Join(tempDir, "other", "file.txt")

	if !handler.IsFileIgnored(gitignoreContent, ignoredFile, tempDir) {
		t.Error("Expected file in secrets/ to be ignored")
	}

	if handler.IsFileIgnored(gitignoreContent, notIgnoredFile, tempDir) {
		t.Error("Expected file not in ignored patterns to not be ignored")
	}
}

func TestGitHandler_AddToGitignore(t *testing.T) {
	handler := NewGitHandler()

	tempDir := t.TempDir()
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	existingContent := "*.log\n"
	filePath := filepath.Join(tempDir, "secrets", "newfile.txt")

	os.WriteFile(gitignorePath, []byte(existingContent), 0644)
	os.MkdirAll(filepath.Dir(filePath), 0755)

	err := handler.AddToGitignore(tempDir, filePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Expected no error reading gitignore, got %v", err)
	}

	expected := existingContent + "\nsecrets/newfile.txt\n"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}
