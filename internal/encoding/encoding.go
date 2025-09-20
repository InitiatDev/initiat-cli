package encoding

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
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
	// Try RawURLEncoding first (our preferred format)
	if data, err := base64.RawURLEncoding.DecodeString(encoded); err == nil {
		return data, nil
	}

	// Try standard base64 encoding (with padding)
	if data, err := base64.StdEncoding.DecodeString(encoded); err == nil {
		return data, nil
	}

	// Try URLEncoding (with padding)
	if data, err := base64.URLEncoding.DecodeString(encoded); err == nil {
		return data, nil
	}

	// Try RawStdEncoding (without padding)
	return base64.RawStdEncoding.DecodeString(encoded)
}
