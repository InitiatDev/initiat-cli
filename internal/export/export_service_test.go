package export

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportService_ExportSecret_NewFile(t *testing.T) {
	tempDir := t.TempDir()
	service := NewExportService(false)
	filePath := filepath.Join(tempDir, "secrets", "newfile.txt")

	err := service.ExportSecret("API_KEY", "secret-value", filePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Expected no error reading file, got %v", err)
	}

	expected := "API_KEY=secret-value\n"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestExportService_ExportSecret_ExistingFile_NewKey(t *testing.T) {
	tempDir := t.TempDir()
	service := NewExportService(false)
	filePath := filepath.Join(tempDir, "secrets.txt")

	existingContent := "EXISTING_KEY=existing-value\n"
	os.WriteFile(filePath, []byte(existingContent), 0644)

	err := service.ExportSecret("NEW_KEY", "new-value", filePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Expected no error reading file, got %v", err)
	}

	expected := existingContent + "NEW_KEY=new-value"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestExportService_ExportSecret_ForceOverride(t *testing.T) {
	tempDir := t.TempDir()
	service := NewExportService(true)
	filePath := filepath.Join(tempDir, "secrets.txt")

	existingContent := "API_KEY=old-value\nOTHER_KEY=other-value\n"
	os.WriteFile(filePath, []byte(existingContent), 0644)

	err := service.ExportSecret("API_KEY", "new-value", filePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Expected no error reading file, got %v", err)
	}

	expected := "API_KEY=new-value\nOTHER_KEY=other-value\n"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestExportService_ExportSecret_CreatesDirectories(t *testing.T) {
	tempDir := t.TempDir()
	service := NewExportService(false)
	filePath := filepath.Join(tempDir, "deep", "nested", "path", "secrets.txt")

	err := service.ExportSecret("API_KEY", "secret-value", filePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist", filePath)
	}
}
