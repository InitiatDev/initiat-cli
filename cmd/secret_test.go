package cmd

import (
	"crypto/rand"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/crypto"
)

func TestEncryptSecretValue(t *testing.T) {
	workspaceKey := make([]byte, crypto.WorkspaceKeySize)
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

	if len(nonce) != crypto.SecretboxNonceSize {
		t.Errorf("Expected nonce size %d, got %d", crypto.SecretboxNonceSize, len(nonce))
	}

	if string(encryptedValue) == testValue {
		t.Error("Encrypted value should be different from original value")
	}
}

func TestEncryptSecretValueDifferentNonces(t *testing.T) {
	workspaceKey := make([]byte, crypto.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		t.Fatalf("Failed to generate test workspace key: %v", err)
	}

	testValue := "test-secret-value"

	_, nonce1, err := encryptSecretValue(testValue, workspaceKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value (first): %v", err)
	}

	_, nonce2, err := encryptSecretValue(testValue, workspaceKey)
	if err != nil {
		t.Fatalf("Failed to encrypt secret value (second): %v", err)
	}

	if string(nonce1) == string(nonce2) {
		t.Error("Nonces should be different for each encryption")
	}
}

func TestEncryptSecretValueEmptyValue(t *testing.T) {
	workspaceKey := make([]byte, crypto.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		t.Fatalf("Failed to generate test workspace key: %v", err)
	}

	encryptedValue, nonce, err := encryptSecretValue("", workspaceKey)
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
