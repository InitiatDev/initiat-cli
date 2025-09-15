package encoding

import (
	"encoding/base64"
	"fmt"
)

// URL-Safe Base64 encoder without padding as per InitFlow encoding specification
var urlSafeNoPadding = base64.URLEncoding.WithPadding(base64.NoPadding)

// Encode encodes binary data to URL-Safe Base64 without padding
func Encode(data []byte) string {
	return urlSafeNoPadding.EncodeToString(data)
}

// Decode decodes URL-Safe Base64 without padding to binary data
func Decode(encoded string) ([]byte, error) {
	return urlSafeNoPadding.DecodeString(encoded)
}

// ValidateAndDecode validates the encoded string format and decodes it
func ValidateAndDecode(encoded string, expectedSize int) ([]byte, error) {
	// Validate no padding characters first
	for _, char := range encoded {
		if char == '=' {
			return nil, fmt.Errorf("encoded string must not contain padding")
		}
	}

	// Validate character set - only A-Z, a-z, 0-9, -, _ allowed
	for _, char := range encoded {
		if !isValidBase64URLChar(char) {
			return nil, fmt.Errorf("invalid character in encoded string")
		}
	}

	// Decode the string
	decoded, err := Decode(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid encoding format")
	}

	// Validate expected size
	if len(decoded) != expectedSize {
		return nil, fmt.Errorf("decoded data has incorrect length: expected %d bytes, got %d", expectedSize, len(decoded))
	}

	return decoded, nil
}

// isValidBase64URLChar checks if a character is valid for URL-Safe Base64
func isValidBase64URLChar(char rune) bool {
	return (char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_'
}

// Data type size constants as per specification
const (
	DeviceIDSize         = 16 // Device IDs: 16 bytes
	UserTokenSize        = 32 // User tokens: 32 bytes
	SessionTokenSize     = 32 // Session tokens: 32 bytes
	Ed25519SignatureSize = 64 // Ed25519 signatures: 64 bytes
	Ed25519PublicKeySize = 32 // Ed25519 public keys: 32 bytes
	X25519PublicKeySize  = 32 // X25519 public keys: 32 bytes
	WorkspaceKeyIDSize   = 16 // Workspace key IDs: 16 bytes
	WorkspaceKeySize     = 32 // Workspace keys: 32 bytes
	X25519PrivateKeySize = 32 // X25519 private keys: 32 bytes
	ChaCha20NonceSize    = 12 // ChaCha20 nonce: 12 bytes
)

// Type-specific encoding functions with validation

// EncodeDeviceID encodes a device ID (16 bytes)
func EncodeDeviceID(deviceID []byte) (string, error) {
	if len(deviceID) != DeviceIDSize {
		return "", fmt.Errorf("device ID must be %d bytes, got %d", DeviceIDSize, len(deviceID))
	}
	return Encode(deviceID), nil
}

// DecodeDeviceID decodes and validates a device ID
func DecodeDeviceID(encoded string) ([]byte, error) {
	return ValidateAndDecode(encoded, DeviceIDSize)
}

// EncodeUserToken encodes a user token (32 bytes)
func EncodeUserToken(token []byte) (string, error) {
	if len(token) != UserTokenSize {
		return "", fmt.Errorf("user token must be %d bytes, got %d", UserTokenSize, len(token))
	}
	return Encode(token), nil
}

// DecodeUserToken decodes and validates a user token
func DecodeUserToken(encoded string) ([]byte, error) {
	return ValidateAndDecode(encoded, UserTokenSize)
}

// EncodeEd25519Signature encodes an Ed25519 signature (64 bytes)
func EncodeEd25519Signature(signature []byte) (string, error) {
	if len(signature) != Ed25519SignatureSize {
		return "", fmt.Errorf("Ed25519 signature must be %d bytes, got %d", Ed25519SignatureSize, len(signature))
	}
	return Encode(signature), nil
}

// DecodeEd25519Signature decodes and validates an Ed25519 signature
func DecodeEd25519Signature(encoded string) ([]byte, error) {
	return ValidateAndDecode(encoded, Ed25519SignatureSize)
}

// EncodeEd25519PublicKey encodes an Ed25519 public key (32 bytes)
func EncodeEd25519PublicKey(publicKey []byte) (string, error) {
	if len(publicKey) != Ed25519PublicKeySize {
		return "", fmt.Errorf("Ed25519 public key must be %d bytes, got %d", Ed25519PublicKeySize, len(publicKey))
	}
	return Encode(publicKey), nil
}

// DecodeEd25519PublicKey decodes and validates an Ed25519 public key
func DecodeEd25519PublicKey(encoded string) ([]byte, error) {
	return ValidateAndDecode(encoded, Ed25519PublicKeySize)
}

// EncodeX25519PublicKey encodes an X25519 public key (32 bytes)
func EncodeX25519PublicKey(publicKey []byte) (string, error) {
	if len(publicKey) != X25519PublicKeySize {
		return "", fmt.Errorf("X25519 public key must be %d bytes, got %d", X25519PublicKeySize, len(publicKey))
	}
	return Encode(publicKey), nil
}

// DecodeX25519PublicKey decodes and validates an X25519 public key
func DecodeX25519PublicKey(encoded string) ([]byte, error) {
	return ValidateAndDecode(encoded, X25519PublicKeySize)
}
