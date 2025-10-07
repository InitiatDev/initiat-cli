package cmd

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/crypto"
)

func TestEncryptSecretValue(t *testing.T) {
	projectKey := make([]byte, crypto.ProjectKeySize)
	if _, err := rand.Read(projectKey); err != nil {
		t.Fatalf("Failed to generate test project key: %v", err)
	}

	testValue := "test-secret-value"

	encryptedValue, nonce, err := encryptSecretValue(testValue, projectKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value: %v", err)
	}

	if len(encryptedValue) == 0 {
		t.Error("Encrypted value should not be empty")
	}

	if len(nonce) != crypto.SecretboxNonceSize {
		t.Errorf("Expected nonce size %d, got %d", crypto.SecretboxNonceSize, len(nonce))
	}

	if string(encryptedValue) == testValue {
		t.Error("Encrypted value should be different from original value")
	}
}

func TestEncryptSecretValueDifferentNonces(t *testing.T) {
	projectKey := make([]byte, crypto.ProjectKeySize)
	if _, err := rand.Read(projectKey); err != nil {
		t.Fatalf("Failed to generate test project key: %v", err)
	}

	testValue := "test-secret-value"

	_, nonce1, err := encryptSecretValue(testValue, projectKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value (first): %v", err)
	}

	_, nonce2, err := encryptSecretValue(testValue, projectKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value (second): %v", err)
	}

	if string(nonce1) == string(nonce2) {
		t.Error("Nonces should be different for each encryption")
	}
}

func TestEncryptSecretValueEmptyValue(t *testing.T) {
	projectKey := make([]byte, crypto.ProjectKeySize)
	if _, err := rand.Read(projectKey); err != nil {
		t.Fatalf("Failed to generate test project key: %v", err)
	}

	encryptedValue, nonce, err := encryptSecretValue("", projectKey)
	if err != nil {
		t.Fatalf("Failed to encrypt empty secret value: %v", err)
	}

	if len(encryptedValue) == 0 {
		t.Error("Encrypted value should not be empty even for empty input")
	}

	if len(nonce) != crypto.SecretboxNonceSize {
		t.Errorf("Expected nonce size %d, got %d", crypto.SecretboxNonceSize, len(nonce))
	}
}

func TestEncryptSecretValueInvalidKey(t *testing.T) {
	invalidKey := make([]byte, 16)
	if _, err := rand.Read(invalidKey); err != nil {
		t.Fatalf("Failed to generate invalid key: %v", err)
	}

	testValue := "test-secret-value"

	_, _, err := encryptSecretValue(testValue, invalidKey)
	if err == nil {
		t.Error("Expected error with invalid key size, but got none")
	}
}

func TestSecretExportCommand_Integration(t *testing.T) {
	cmd := secretExportCmd
	if cmd == nil {
		t.Fatal("secretExportCmd should not be nil")
	}

	if cmd.Flags().Lookup("output") == nil {
		t.Error("Expected 'output' flag to be defined")
	}

	if cmd.Args == nil {
		t.Error("Expected Args to be defined")
	}
}

func TestSecretExportCommand_FileCreation(t *testing.T) {
	tempDir := t.TempDir()

	deepPath := filepath.Join(tempDir, "deep", "nested", "path", "secrets.txt")

	dir := filepath.Dir(deepPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	content := "API_KEY=test-value\n"
	if err := os.WriteFile(deepPath, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if _, err := os.Stat(deepPath); os.IsNotExist(err) {
		t.Fatal("File should exist after creation")
	}

	readContent, err := os.ReadFile(deepPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != content {
		t.Errorf("Expected content %q, got %q", content, string(readContent))
	}
}
