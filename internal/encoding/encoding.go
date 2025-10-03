package encoding

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	WorkspaceKeySize     = 32 // 256 bits for XSalsa20Poly1305 (NaCl secretbox)
	X25519PrivateKeySize = 32 // X25519 private key size
	SecretboxNonceSize   = 24 // XSalsa20Poly1305 nonce size (NaCl secretbox)
	UserTokenSize        = 32 // User authentication token size
)

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

func WrapWorkspaceKey(workspaceKey []byte, devicePublicKey []byte) (string, error) {
	if len(workspaceKey) != WorkspaceKeySize {
		return "", fmt.Errorf("invalid workspace key size: %d", len(workspaceKey))
	}

	if len(devicePublicKey) != X25519PrivateKeySize {
		return "", fmt.Errorf("invalid device public key size: %d", len(devicePublicKey))
	}

	// Generate ephemeral private key (same as original workspace key wrapping)
	ephemeralPrivate := make([]byte, X25519PrivateKeySize)
	if _, err := rand.Read(ephemeralPrivate); err != nil {
		return "", fmt.Errorf("failed to generate ephemeral private key: %w", err)
	}

	// Generate ephemeral public key
	ephemeralPublic, err := curve25519.X25519(ephemeralPrivate, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral public key: %w", err)
	}

	// Compute shared secret using device's public key (instead of admin's private key)
	sharedSecret, err := curve25519.X25519(ephemeralPrivate, devicePublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to compute shared secret: %w", err)
	}

	// Use HKDF to derive encryption key (same as original)
	hkdf := hkdf.New(sha256.New, sharedSecret, []byte("initiat.wrap"), []byte("workspace"))
	encryptionKey := make([]byte, WorkspaceKeySize)
	if _, err := hkdf.Read(encryptionKey); err != nil {
		return "", fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Use ChaCha20-Poly1305 (same as original)
	cipher, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate nonce for ChaCha20-Poly1305
	const chacha20NonceSize = 12
	nonce := make([]byte, chacha20NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt using ChaCha20-Poly1305 with random nonce
	ciphertext := cipher.Seal(nil, nonce, workspaceKey, nil) // #nosec G407 - nonce is randomly generated

	// Create wrapped key with ephemeral public key + nonce + ciphertext
	wrapped := make([]byte, 0, 32+12+len(ciphertext))
	wrapped = append(wrapped, ephemeralPublic...)
	wrapped = append(wrapped, nonce...)
	wrapped = append(wrapped, ciphertext...)

	return Encode(wrapped), nil
}
