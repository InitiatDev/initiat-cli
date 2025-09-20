package cmd

import (
	"crypto/rand"
	"testing"

	"github.com/DylanBlakemore/initflow-cli/internal/encoding"
)

func TestEncryptSecretValue(t *testing.T) {
	workspaceKey := make([]byte, encoding.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		t.Fatalf("Failed to generate test workspace key: %v", err)
	}

	testValue := "test-secret-value"

	encryptedValue, nonce, err := encryptSecretValue(testValue, workspaceKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value: %v", err)
	}

	if len(encryptedValue) == 0 {
		t.Error("Encrypted value should not be empty")
	}

	// Verify nonce is correct size
	if len(nonce) != encoding.SecretboxNonceSize {
		t.Errorf("Expected nonce size %d, got %d", encoding.SecretboxNonceSize, len(nonce))
	}

	// Verify encrypted value is different from original
	if string(encryptedValue) == testValue {
		t.Error("Encrypted value should be different from original value")
	}
}

func TestEncryptSecretValueDifferentNonces(t *testing.T) {
	// Generate a test workspace key
	workspaceKey := make([]byte, encoding.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		t.Fatalf("Failed to generate test workspace key: %v", err)
	}

	testValue := "test-secret-value"

	// Encrypt the same value twice
	_, nonce1, err := encryptSecretValue(testValue, workspaceKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value (first): %v", err)
	}

	_, nonce2, err := encryptSecretValue(testValue, workspaceKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value (second): %v", err)
	}

	// Verify nonces are different (extremely unlikely to be the same)
	if string(nonce1) == string(nonce2) {
		t.Error("Nonces should be different for each encryption")
	}
}

func TestEncryptSecretValueEmptyValue(t *testing.T) {
	// Generate a test workspace key
	workspaceKey := make([]byte, encoding.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		t.Fatalf("Failed to generate test workspace key: %v", err)
	}

	// Test with empty value
	encryptedValue, nonce, err := encryptSecretValue("", workspaceKey)
	if err != nil {
		t.Fatalf("Failed to encrypt empty secret value: %v", err)
	}

	// Should still produce valid encrypted output
	if len(encryptedValue) == 0 {
		t.Error("Encrypted value should not be empty even for empty input")
	}

	if len(nonce) != encoding.SecretboxNonceSize {
		t.Errorf("Expected nonce size %d, got %d", encoding.SecretboxNonceSize, len(nonce))
	}
}

func TestEncryptSecretValueInvalidKey(t *testing.T) {
	// Test with invalid workspace key size
	invalidKey := make([]byte, 16) // Wrong size, should be 32
	if _, err := rand.Read(invalidKey); err != nil {
		t.Fatalf("Failed to generate invalid key: %v", err)
	}

	testValue := "test-secret-value"

	// Should fail with invalid key size
	_, _, err := encryptSecretValue(testValue, invalidKey)
	if err == nil {
		t.Error("Expected error with invalid key size, but got none")
	}
}
