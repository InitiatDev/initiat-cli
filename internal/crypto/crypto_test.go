package crypto

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/curve25519"
)

func TestGenerateEd25519Keypair(t *testing.T) {
	publicKey, privateKey, err := GenerateEd25519Keypair()
	assert.NoError(t, err)
	assert.Equal(t, ed25519.PublicKeySize, len(publicKey))
	assert.Equal(t, ed25519.PrivateKeySize, len(privateKey))

	// Test that the keypair works for signing/verification
	message := []byte("test message")
	signature := ed25519.Sign(privateKey, message)
	assert.True(t, ed25519.Verify(publicKey, message, signature))
}

func TestGenerateX25519Keypair(t *testing.T) {
	publicKey, privateKey, err := GenerateX25519Keypair()
	assert.NoError(t, err)
	assert.Equal(t, X25519PrivateKeySize, len(publicKey))
	assert.Equal(t, X25519PrivateKeySize, len(privateKey))

	// Test that the keypair works for key exchange
	expectedPublicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	assert.NoError(t, err)
	assert.Equal(t, expectedPublicKey, publicKey)
}

func TestEncryptDecryptSecretValue(t *testing.T) {
	projectKey := make([]byte, ProjectKeySize)
	for i := range projectKey {
		projectKey[i] = byte(i)
	}

	originalValue := "test-secret-value"

	// Encrypt
	ciphertext, nonce, err := EncryptSecretValue(originalValue, projectKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.Equal(t, SecretboxNonceSize, len(nonce))

	// Decrypt
	decryptedValue, err := DecryptSecretValue(ciphertext, nonce, projectKey)
	assert.NoError(t, err)
	assert.Equal(t, originalValue, decryptedValue)
}

func TestEncryptSecretValueDifferentNonces(t *testing.T) {
	projectKey := make([]byte, ProjectKeySize)
	for i := range projectKey {
		projectKey[i] = byte(i)
	}

	originalValue := "test-secret-value"

	_, nonce1, err := EncryptSecretValue(originalValue, projectKey)
	assert.NoError(t, err)

	_, nonce2, err := EncryptSecretValue(originalValue, projectKey)
	assert.NoError(t, err)

	// Nonces should be different
	assert.NotEqual(t, nonce1, nonce2)
}

func TestEncryptSecretValueInvalidKey(t *testing.T) {
	invalidKey := make([]byte, 16) // Wrong size
	_, _, err := EncryptSecretValue("test", invalidKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project key size")
}

func TestDecryptSecretValueInvalidKey(t *testing.T) {
	invalidKey := make([]byte, 16) // Wrong size
	_, err := DecryptSecretValue([]byte("test"), []byte("test"), invalidKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project key size")
}

func TestDecryptSecretValueInvalidNonce(t *testing.T) {
	projectKey := make([]byte, ProjectKeySize)
	invalidNonce := make([]byte, 16) // Wrong size
	_, err := DecryptSecretValue([]byte("test"), invalidNonce, projectKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nonce size")
}
