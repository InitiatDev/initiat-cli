package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHandler_FindGitRoot(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	gitDir := filepath.Join(tempDir, "project", ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	subDir := filepath.Join(tempDir, "project", "sub", "dir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	root, found := handler.FindGitRoot(subDir)
	if !found {
		t.Errorf("Should find git root")
	}

	expectedRoot := filepath.Join(tempDir, "project")
	if root != expectedRoot {
		t.Errorf("Expected git root %q, got %q", expectedRoot, root)
	}
}

func TestHandler_FindGitRoot_NotFound(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	dir := filepath.Join(tempDir, "notgit")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	_, found := handler.FindGitRoot(dir)
	if found {
		t.Errorf("Should not find git root in non-git directory")
	}
}

func TestHandler_IsGitRepository(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	nonGitDir := filepath.Join(tempDir, "notgit")
	err := os.MkdirAll(nonGitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if handler.IsGitRepository(nonGitDir) {
		t.Errorf("Should not be a git repository")
	}

	gitDir := filepath.Join(tempDir, "gitproject")
	err = os.MkdirAll(filepath.Join(gitDir, ".git"), 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	if !handler.IsGitRepository(gitDir) {
		t.Errorf("Should be a git repository")
	}
}

func TestHandler_ReadGitignore(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	gitignoreContent := "*.log\n*.tmp\n"
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	content, err := handler.ReadGitignore(tempDir)
	if err != nil {
		t.Fatalf("ReadGitignore failed: %v", err)
	}

	if content != gitignoreContent {
		t.Errorf("Expected content %q, got %q", gitignoreContent, content)
	}
}

func TestHandler_ReadGitignore_NotExists(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	content, err := handler.ReadGitignore(tempDir)
	if err != nil {
		t.Fatalf("ReadGitignore failed: %v", err)
	}

	if content != "" {
		t.Errorf("Expected empty content, got %q", content)
	}
}

func TestHandler_ReadFile(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "test.txt")
	content := "Hello, World!"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	readContent, err := handler.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}
}

func TestHandler_WriteFile(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "test.txt")
	content := "Hello, World!"

	err := handler.WriteFile(filePath, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(readContent) != content {
		t.Errorf("Expected content %q, got %q", content, string(readContent))
	}
}

func TestHandler_IsFileIgnored(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	gitignoreContent := `*.log
*.tmp
build/
src/test/
`

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "ignored by pattern",
			filePath: filepath.Join(tempDir, "app.log"),
			expected: true,
		},
		{
			name:     "ignored by directory",
			filePath: filepath.Join(tempDir, "build", "output"),
			expected: true,
		},
		{
			name:     "ignored by nested directory",
			filePath: filepath.Join(tempDir, "src", "test", "test.go"),
			expected: true,
		},
		{
			name:     "not ignored",
			filePath: filepath.Join(tempDir, "src", "main.go"),
			expected: false,
		},
		{
			name:     "not ignored file",
			filePath: filepath.Join(tempDir, "config.txt"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsFileIgnored(gitignoreContent, tt.filePath, tempDir)
			if result != tt.expected {
				t.Errorf("IsFileIgnored() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHandler_AddToGitignore(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	initialContent := "*.log\n"
	err := os.WriteFile(gitignorePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial .gitignore: %v", err)
	}

	filePath := filepath.Join(tempDir, "secrets", "config.env")
	err = handler.AddToGitignore(tempDir, filePath)
	if err != nil {
		t.Fatalf("AddToGitignore failed: %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read updated .gitignore: %v", err)
	}

	expectedContent := initialContent + "secrets/config.env\n"
	if string(content) != expectedContent {
		t.Errorf("Expected .gitignore content %q, got %q", expectedContent, string(content))
	}
}

func TestHandler_AddToGitignore_NoInitialGitignore(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "secrets", "config.env")
	err := handler.AddToGitignore(tempDir, filePath)
	if err != nil {
		t.Fatalf("AddToGitignore failed: %v", err)
	}

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read created .gitignore: %v", err)
	}

	expectedContent := "secrets/config.env\n"
	if string(content) != expectedContent {
		t.Errorf("Expected .gitignore content %q, got %q", expectedContent, string(content))
	}
}
