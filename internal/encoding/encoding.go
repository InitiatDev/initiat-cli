package encoding

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
)

const (
	WorkspaceKeySize     = 32 // 256 bits for ChaCha20-Poly1305
	X25519PrivateKeySize = 32 // X25519 private key size
	ChaCha20NonceSize    = 12 // ChaCha20-Poly1305 nonce size
	UserTokenSize        = 32 // User authentication token size
)

// EncodeEd25519PublicKey encodes an Ed25519 public key for API transmission
func EncodeEd25519PublicKey(publicKey ed25519.PublicKey) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid Ed25519 public key size: %d", len(publicKey))
	}
	return base64.StdEncoding.EncodeToString(publicKey), nil
}

// EncodeX25519PublicKey encodes an X25519 public key for API transmission
func EncodeX25519PublicKey(publicKey []byte) (string, error) {
	if len(publicKey) != UserTokenSize {
		return "", fmt.Errorf("invalid X25519 public key size: %d", len(publicKey))
	}
	return base64.StdEncoding.EncodeToString(publicKey), nil
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
func Decode(encoded string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(encoded)
}
