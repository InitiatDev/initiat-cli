package encoding

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/curve25519"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte{0x42}},
		{"multiple bytes", []byte{0x01, 0x02, 0x03, 0x04}},
		{"random bytes", []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Encode(tt.input)
			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
			}
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("Round-trip failed: got %v, expected %v", decoded, tt.input)
			}
		})
	}
}

func TestEncodeEd25519PublicKey(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := EncodeEd25519PublicKey(publicKey)
	if err != nil {
		t.Errorf("EncodeEd25519PublicKey() error = %v", err)
	}

	if encoded == "" {
		t.Error("Encoded public key is empty")
	}
}

func TestEncodeX25519PublicKey(t *testing.T) {
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	encoded, err := EncodeX25519PublicKey(publicKey)
	if err != nil {
		t.Errorf("EncodeX25519PublicKey() error = %v", err)
	}

	if encoded == "" {
		t.Error("Encoded public key is empty")
	}
}

func TestEncodeEd25519Signature(t *testing.T) {
	signature := make([]byte, ed25519.SignatureSize)
	rand.Read(signature)

	encoded, err := EncodeEd25519Signature(signature)
	if err != nil {
		t.Errorf("EncodeEd25519Signature() error = %v", err)
	}

	if encoded == "" {
		t.Error("Encoded signature is empty")
	}
}

func TestWrapWorkspaceKey(t *testing.T) {
	workspaceKey := make([]byte, WorkspaceKeySize)
	for i := range workspaceKey {
		workspaceKey[i] = byte(i % 256)
	}

	devicePrivateKey := make([]byte, X25519PrivateKeySize)
	for i := range devicePrivateKey {
		devicePrivateKey[i] = byte(i % 256)
	}
	devicePublicKey, err := curve25519.X25519(devicePrivateKey, curve25519.Basepoint)
	if err != nil {
		t.Fatalf("Failed to generate device public key: %v", err)
	}

	wrapped, err := WrapWorkspaceKey(workspaceKey, devicePublicKey)
	if err != nil {
		t.Fatalf("WrapWorkspaceKey failed: %v", err)
	}

	if wrapped == "" {
		t.Error("Wrapped key should not be empty")
	}

	wrapped2, err := WrapWorkspaceKey(workspaceKey, devicePublicKey)
	if err != nil {
		t.Fatalf("WrapWorkspaceKey failed on second call: %v", err)
	}

	if wrapped == wrapped2 {
		t.Error("Wrapped keys should be different due to ephemeral key generation")
	}
}

func TestWrapWorkspaceKeyInvalidInput(t *testing.T) {
	devicePrivateKey := make([]byte, X25519PrivateKeySize)
	rand.Read(devicePrivateKey)
	devicePublicKey, _ := curve25519.X25519(devicePrivateKey, curve25519.Basepoint)

	tests := []struct {
		name         string
		workspaceKey []byte
		deviceKey    []byte
		expectError  bool
	}{
		{
			name:         "invalid workspace key size",
			workspaceKey: make([]byte, 16),
			deviceKey:    devicePublicKey,
			expectError:  true,
		},
		{
			name:         "invalid device key size",
			workspaceKey: make([]byte, WorkspaceKeySize),
			deviceKey:    make([]byte, 16),
			expectError:  true,
		},
		{
			name:         "valid inputs",
			workspaceKey: make([]byte, WorkspaceKeySize),
			deviceKey:    devicePublicKey,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WrapWorkspaceKey(tt.workspaceKey, tt.deviceKey)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
