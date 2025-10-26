package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandler_EnsureDirectory(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	dirPath := filepath.Join(tempDir, "test", "nested", "dir")
	err := handler.EnsureDirectory(dirPath)
	if err != nil {
		t.Fatalf("EnsureDirectory failed: %v", err)
	}

	parentDir := filepath.Dir(dirPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Parent directory was not created: %s", parentDir)
	}
}

func TestHandler_CreateDirectory(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	dirPath := filepath.Join(tempDir, "testdir")
	err := handler.CreateDirectory(dirPath)
	if err != nil {
		t.Fatalf("CreateDirectory failed: %v", err)
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", dirPath)
	}
}

func TestHandler_ReadWriteFile(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")
	content := "Hello, World!"

	err := handler.WriteFile(filePath, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	readContent, err := handler.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}
}

func TestHandler_FileExists(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	if handler.FileExists(filePath) {
		t.Errorf("FileExists should return false for non-existent file")
	}

	err := handler.WriteFile(filePath, "test content")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if !handler.FileExists(filePath) {
		t.Errorf("FileExists should return true for existing file")
	}
}

func TestHandler_RemoveFile(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	err := handler.WriteFile(filePath, "test content")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if !handler.FileExists(filePath) {
		t.Fatalf("File should exist before removal")
	}

	err = handler.RemoveFile(filePath)
	if err != nil {
		t.Fatalf("RemoveFile failed: %v", err)
	}

	if handler.FileExists(filePath) {
		t.Errorf("File should not exist after removal")
	}
}

func TestHandler_CreateReadSymlink(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	targetPath := filepath.Join(tempDir, "target")
	err := handler.CreateDirectory(targetPath)
	if err != nil {
		t.Fatalf("CreateDirectory failed: %v", err)
	}

	linkPath := filepath.Join(tempDir, "link")
	err = handler.CreateSymlink(targetPath, linkPath)
	if err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	readTarget, err := handler.ReadSymlink(linkPath)
	if err != nil {
		t.Fatalf("ReadSymlink failed: %v", err)
	}

	if readTarget != targetPath {
		t.Errorf("Expected symlink target %q, got %q", targetPath, readTarget)
	}
}

func TestHandler_GetWorkingDirectory(t *testing.T) {
	handler := NewHandler()

	wd, err := handler.GetWorkingDirectory()
	if err != nil {
		t.Fatalf("GetWorkingDirectory failed: %v", err)
	}

	if wd == "" {
		t.Errorf("Working directory should not be empty")
	}
}

func TestHandler_JoinPaths(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "single path",
			paths:    []string{"test"},
			expected: "test",
		},
		{
			name:     "multiple paths",
			paths:    []string{"dir1", "dir2", "file.txt"},
			expected: filepath.Join("dir1", "dir2", "file.txt"),
		},
		{
			name:     "empty paths",
			paths:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.JoinPaths(tt.paths...)
			if result != tt.expected {
				t.Errorf("JoinPaths() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHandler_GetAbsolutePath(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	absPath := tempDir
	result, err := handler.GetAbsolutePath(absPath)
	if err != nil {
		t.Fatalf("GetAbsolutePath failed: %v", err)
	}
	if result != absPath {
		t.Errorf("Expected %q, got %q", absPath, result)
	}

	relPath := "test.txt"
	result, err = handler.GetAbsolutePath(relPath)
	if err != nil {
		t.Fatalf("GetAbsolutePath failed: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("Result should be absolute path: %q", result)
	}
}

func TestHandler_GetRelativePath(t *testing.T) {
	handler := NewHandler()
	tempDir := t.TempDir()

	basePath := filepath.Join(tempDir, "base")
	targetPath := filepath.Join(tempDir, "base", "sub", "file.txt")

	expected, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		t.Fatalf("filepath.Rel failed: %v", err)
	}

	result, err := handler.GetRelativePath(basePath, targetPath)
	if err != nil {
		t.Fatalf("GetRelativePath failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestHandler_GetDirectory(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "file path",
			path:     "/path/to/file.txt",
			expected: "/path/to",
		},
		{
			name:     "directory path",
			path:     "/path/to/dir",
			expected: "/path/to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GetDirectory(tt.path)
			if result != tt.expected {
				t.Errorf("GetDirectory() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHandler_GetProjectPath(t *testing.T) {
	handler := NewHandler()

	projectDir := "testproject"
	result, err := handler.GetProjectPath(projectDir)
	if err != nil {
		t.Fatalf("GetProjectPath failed: %v", err)
	}

	expected := filepath.Join(".", projectDir)
	if !strings.HasSuffix(result, expected) {
		t.Errorf("Expected path to end with %q, got %q", expected, result)
	}
}

func TestHandler_GetSubPath(t *testing.T) {
	handler := NewHandler()

	basePath := "/base"
	subDir := "sub"

	result, err := handler.GetSubPath(basePath, subDir)
	if err != nil {
		t.Fatalf("GetSubPath failed: %v", err)
	}

	expected := filepath.Join(basePath, subDir)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestHandler_GetFilePath(t *testing.T) {
	handler := NewHandler()

	basePath := "/base"
	filename := "file.txt"

	result, err := handler.GetFilePath(basePath, filename)
	if err != nil {
		t.Fatalf("GetFilePath failed: %v", err)
	}

	expected := filepath.Join(basePath, filename)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestHandler_FindKeyInContent(t *testing.T) {
	handler := NewHandler()

	content := `KEY1=value1
KEY2=value2
KEY3=value3`

	tests := []struct {
		name     string
		key      string
		expected int
		found    bool
	}{
		{
			name:     "existing key",
			key:      "KEY2",
			expected: 1,
			found:    true,
		},
		{
			name:     "non-existing key",
			key:      "NONEXISTENT",
			expected: -1,
			found:    false,
		},
		{
			name:     "key with spaces",
			key:      "KEY1",
			expected: 0,
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, found := handler.FindKeyInContent(content, tt.key)
			if found != tt.found {
				t.Errorf("FindKeyInContent() found = %v, want %v", found, tt.found)
			}
			if found && index != tt.expected {
				t.Errorf("FindKeyInContent() index = %v, want %v", index, tt.expected)
			}
		})
	}
}

func TestHandler_UpdateKeyInContent(t *testing.T) {
	handler := NewHandler()

	content := `KEY1=value1
KEY2=value2
KEY3=value3`

	result := handler.UpdateKeyInContent(content, "KEY2", "newvalue", 1)

	expected := `KEY1=value1
KEY2=newvalue
KEY3=value3`

	if result != expected {
		t.Errorf("UpdateKeyInContent() = %v, want %v", result, expected)
	}
}

func TestHandler_AppendKeyToContent(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		name     string
		content  string
		key      string
		value    string
		expected string
	}{
		{
			name:     "empty content",
			content:  "",
			key:      "NEWKEY",
			value:    "newvalue",
			expected: "NEWKEY=newvalue\n",
		},
		{
			name:     "content with newline",
			content:  "KEY1=value1\n",
			key:      "NEWKEY",
			value:    "newvalue",
			expected: "KEY1=value1\nNEWKEY=newvalue",
		},
		{
			name:     "content without newline",
			content:  "KEY1=value1",
			key:      "NEWKEY",
			value:    "newvalue",
			expected: "KEY1=value1\nNEWKEY=newvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.AppendKeyToContent(tt.content, tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("AppendKeyToContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
