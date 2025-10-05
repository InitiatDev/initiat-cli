package export

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileHandler_EnsureDirectory(t *testing.T) {
	handler := NewFileHandler()

	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "deep", "nested", "path", "file.txt")

	err := handler.EnsureDirectory(testPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	dir := filepath.Dir(testPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("Directory %s should exist", dir)
	}
}

func TestFileHandler_FileExists(t *testing.T) {
	handler := NewFileHandler()

	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	nonExistingFile := filepath.Join(tempDir, "nonexisting.txt")

	os.WriteFile(existingFile, []byte("test"), 0644)

	if !handler.FileExists(existingFile) {
		t.Error("Expected existing file to exist")
	}

	if handler.FileExists(nonExistingFile) {
		t.Error("Expected non-existing file to not exist")
	}
}

func TestFileHandler_ReadWriteFile(t *testing.T) {
	handler := NewFileHandler()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := "test content"

	err := handler.WriteFile(testFile, content)
	if err != nil {
		t.Fatalf("Expected no error writing file, got %v", err)
	}

	readContent, err := handler.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Expected no error reading file, got %v", err)
	}

	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}
}

func TestFileHandler_FindKeyInContent(t *testing.T) {
	handler := NewFileHandler()

	content := "KEY1=value1\nKEY2=value2\nKEY3=value3"

	index, found := handler.FindKeyInContent(content, "KEY2")
	if !found {
		t.Error("Expected to find KEY2")
	}
	if index != 1 {
		t.Errorf("Expected index 1, got %d", index)
	}

	index, found = handler.FindKeyInContent(content, "NONEXISTENT")
	if found {
		t.Error("Expected not to find NONEXISTENT key")
	}
	if index != -1 {
		t.Errorf("Expected index -1, got %d", index)
	}
}

func TestFileHandler_UpdateKeyInContent(t *testing.T) {
	handler := NewFileHandler()

	content := "KEY1=value1\nKEY2=oldvalue\nKEY3=value3"
	expected := "KEY1=value1\nKEY2=newvalue\nKEY3=value3"

	result := handler.UpdateKeyInContent(content, "KEY2", "newvalue", 1)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFileHandler_AppendKeyToContent(t *testing.T) {
	handler := NewFileHandler()

	emptyContent := ""
	result1 := handler.AppendKeyToContent(emptyContent, "NEWKEY", "newvalue")
	expected1 := "NEWKEY=newvalue\n"
	if result1 != expected1 {
		t.Errorf("Expected %q, got %q", expected1, result1)
	}

	existingContent := "KEY1=value1\nKEY2=value2"
	result2 := handler.AppendKeyToContent(existingContent, "NEWKEY", "newvalue")
	expected2 := "KEY1=value1\nKEY2=value2\nNEWKEY=newvalue"
	if result2 != expected2 {
		t.Errorf("Expected %q, got %q", expected2, result2)
	}
}
