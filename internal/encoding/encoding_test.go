package encoding

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
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
