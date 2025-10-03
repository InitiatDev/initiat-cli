package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	WorkspaceKeySize     = 32 // 256 bits for XSalsa20Poly1305 (NaCl secretbox)
	X25519PrivateKeySize = 32 // X25519 private key size
	SecretboxNonceSize   = 24 // XSalsa20Poly1305 nonce size (NaCl secretbox)
	UserTokenSize        = 32 // User authentication token size
)

// GenerateEd25519Keypair generates a new Ed25519 keypair
func GenerateEd25519Keypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 keypair: %w", err)
	}
	return publicKey, privateKey, nil
}

// GenerateX25519Keypair generates a new X25519 keypair
func GenerateX25519Keypair() ([]byte, []byte, error) {
	privateKey := make([]byte, X25519PrivateKeySize)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate X25519 private key: %w", err)
	}

	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate X25519 public key: %w", err)
	}

	return publicKey, privateKey, nil
}

// EncodeEd25519PublicKey encodes an Ed25519 public key for API transmission
func EncodeEd25519PublicKey(publicKey ed25519.PublicKey) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid Ed25519 public key size: %d", len(publicKey))
	}
	return base64.RawURLEncoding.EncodeToString(publicKey), nil
}

// EncodeX25519PublicKey encodes an X25519 public key for API transmission
func EncodeX25519PublicKey(publicKey []byte) (string, error) {
	if len(publicKey) != X25519PrivateKeySize {
		return "", fmt.Errorf("invalid X25519 public key size: %d", len(publicKey))
	}
	return base64.RawURLEncoding.EncodeToString(publicKey), nil
}

// EncodeEd25519Signature encodes an Ed25519 signature for authentication headers
func EncodeEd25519Signature(signature []byte) (string, error) {
	if len(signature) != ed25519.SignatureSize {
		return "", fmt.Errorf("invalid Ed25519 signature size: %d", len(signature))
	}
	return base64.RawURLEncoding.EncodeToString(signature), nil
}

// Encode provides general base64 encoding for binary data (like wrapped keys)
func Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Decode provides general base64 decoding for binary data
// Tries multiple base64 formats for compatibility
func Decode(encoded string) ([]byte, error) {
	if data, err := base64.RawURLEncoding.DecodeString(encoded); err == nil {
		return data, nil
	}

	if data, err := base64.StdEncoding.DecodeString(encoded); err == nil {
		return data, nil
	}

	if data, err := base64.URLEncoding.DecodeString(encoded); err == nil {
		return data, nil
	}

	return base64.RawStdEncoding.DecodeString(encoded)
}

// WrapWorkspaceKey wraps a workspace key with a device public key
func WrapWorkspaceKey(workspaceKey []byte, devicePublicKey []byte) (string, error) {
	if len(workspaceKey) != WorkspaceKeySize {
		return "", fmt.Errorf("invalid workspace key size: %d", len(workspaceKey))
	}

	if len(devicePublicKey) != X25519PrivateKeySize {
		return "", fmt.Errorf("invalid device public key size: %d", len(devicePublicKey))
	}

	ephemeralPrivate := make([]byte, X25519PrivateKeySize)
	if _, err := rand.Read(ephemeralPrivate); err != nil {
		return "", fmt.Errorf("failed to generate ephemeral private key: %w", err)
	}

	ephemeralPublic, err := curve25519.X25519(ephemeralPrivate, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral public key: %w", err)
	}

	sharedSecret, err := curve25519.X25519(ephemeralPrivate, devicePublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to compute shared secret: %w", err)
	}

	hkdf := hkdf.New(sha256.New, sharedSecret, []byte("initiat.wrap"), []byte("workspace"))
	encryptionKey := make([]byte, WorkspaceKeySize)
	if _, err := hkdf.Read(encryptionKey); err != nil {
		return "", fmt.Errorf("failed to derive encryption key: %w", err)
	}

	cipher, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	const chacha20NonceSize = 12
	nonce := make([]byte, chacha20NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := cipher.Seal(nil, nonce, workspaceKey, nil) // #nosec G407 - nonce is randomly generated above

	wrapped := make([]byte, 0, 32+12+len(ciphertext))
	wrapped = append(wrapped, ephemeralPublic...)
	wrapped = append(wrapped, nonce...)
	wrapped = append(wrapped, ciphertext...)

	return Encode(wrapped), nil
}

// UnwrapWorkspaceKey unwraps a workspace key with a device private key
func UnwrapWorkspaceKey(wrappedKey string, devicePrivateKey []byte) ([]byte, error) {
	if len(devicePrivateKey) != X25519PrivateKeySize {
		return nil, fmt.Errorf("invalid device private key size: %d", len(devicePrivateKey))
	}

	wrapped, err := Decode(wrappedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wrapped key: %w", err)
	}

	if len(wrapped) < 32+12 {
		return nil, fmt.Errorf("invalid wrapped key length: %d", len(wrapped))
	}

	ephemeralPublic := wrapped[0:32]
	nonce := wrapped[32:44]
	ciphertext := wrapped[44:]

	sharedSecret, err := curve25519.X25519(devicePrivateKey, ephemeralPublic)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shared secret: %w", err)
	}

	hkdf := hkdf.New(sha256.New, sharedSecret, []byte("initiat.wrap"), []byte("workspace"))
	encryptionKey := make([]byte, WorkspaceKeySize)
	if _, err := hkdf.Read(encryptionKey); err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}

	cipher, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	workspaceKey, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt workspace key: %w", err)
	}

	if len(workspaceKey) != WorkspaceKeySize {
		return nil, fmt.Errorf("invalid workspace key size after unwrap: %d", len(workspaceKey))
	}

	return workspaceKey, nil
}

// EncryptSecretValue encrypts a secret value using ChaCha20Poly1305
func EncryptSecretValue(value string, workspaceKey []byte) ([]byte, []byte, error) {
	if len(workspaceKey) != WorkspaceKeySize {
		return nil, nil, fmt.Errorf(
			"invalid workspace key size: %d bytes, expected %d bytes",
			len(workspaceKey), WorkspaceKeySize)
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	var key [32]byte
	copy(key[:], workspaceKey)

	ciphertext := secretbox.Seal(nil, []byte(value), &nonce, &key)

	return ciphertext, nonce[:], nil
}

// DecryptSecretValue decrypts a secret value using ChaCha20Poly1305
func DecryptSecretValue(ciphertext []byte, nonce []byte, workspaceKey []byte) (string, error) {
	if len(workspaceKey) != WorkspaceKeySize {
		return "", fmt.Errorf(
			"invalid workspace key size: %d bytes, expected %d bytes",
			len(workspaceKey), WorkspaceKeySize)
	}

	if len(nonce) != SecretboxNonceSize {
		return "", fmt.Errorf(
			"invalid nonce size: %d bytes, expected %d bytes",
			len(nonce), SecretboxNonceSize)
	}

	var key [32]byte
	copy(key[:], workspaceKey)

	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &key)
	if !ok {
		return "", fmt.Errorf("failed to decrypt secret value")
	}

	return string(plaintext), nil
}
