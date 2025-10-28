package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestGenerateJSONSchema_ValidJSON(t *testing.T) {
	schema, err := GenerateJSONSchema()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if schemaMap["type"] != "object" {
		t.Errorf("Expected object type, got: %v", schemaMap["type"])
	}
}

func TestSaveSchemaToFile_ContentIntegrity(t *testing.T) {
	originalSchema, err := GenerateJSONSchema()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	tempFile := "/tmp/test-schema.json"
	defer os.Remove(tempFile)

	err = SaveSchemaToFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
	}

	savedSchema, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read saved schema: %v", err)
	}

	if !bytes.Equal(originalSchema, savedSchema) {
		t.Error("Saved schema doesn't match original schema")
	}
}
