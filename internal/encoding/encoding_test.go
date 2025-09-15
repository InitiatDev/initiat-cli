package encoding

import (
	"bytes"
	"strings"
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

			// Verify no padding characters
			if len(encoded) > 0 && encoded[len(encoded)-1] == '=' {
				t.Error("Encoded string contains padding character")
			}

			// Verify only valid characters
			for _, char := range encoded {
				if !isValidBase64URLChar(char) {
					t.Errorf("Invalid character in encoded string: %c", char)
				}
			}

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

func TestValidateAndDecode(t *testing.T) {
	tests := []struct {
		name         string
		encoded      string
		expectedSize int
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "valid device ID",
			encoded:      "ASNFZ4mrze_-3LqYdlQyEA",
			expectedSize: 16,
			wantError:    false,
		},
		{
			name:         "invalid character",
			encoded:      "ASNFZ4mrze/+3LqYdlQyEA", // contains / and +
			expectedSize: 16,
			wantError:    true,
			errorMsg:     "invalid character",
		},
		{
			name:         "contains padding",
			encoded:      "ASNFZ4mrze_-3LqYdlQyEA=",
			expectedSize: 16,
			wantError:    true,
			errorMsg:     "padding",
		},
		{
			name:         "wrong size",
			encoded:      "ASNFZ4mrze_-3LqYdlQyEA",
			expectedSize: 32,
			wantError:    true,
			errorMsg:     "incorrect length",
		},
		{
			name:         "invalid base64",
			encoded:      "!!!invalid!!!",
			expectedSize: 16,
			wantError:    true,
			errorMsg:     "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := ValidateAndDecode(tt.encoded, tt.expectedSize)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !containsIgnoreCase(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(decoded) != tt.expectedSize {
					t.Errorf("Expected %d bytes, got %d", tt.expectedSize, len(decoded))
				}
			}
		})
	}
}

func TestTypeSpecificEncodingFunctions(t *testing.T) {
	t.Run("DeviceID", func(t *testing.T) {
		deviceID := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0xFE, 0xDC, 0xBA, 0x98, 0x76, 0x54, 0x32, 0x10}

		encoded, err := EncodeDeviceID(deviceID)
		if err != nil {
			t.Errorf("EncodeDeviceID() error = %v", err)
		}

		decoded, err := DecodeDeviceID(encoded)
		if err != nil {
			t.Errorf("DecodeDeviceID() error = %v", err)
		}

		if !bytes.Equal(decoded, deviceID) {
			t.Errorf("Round-trip failed: got %v, expected %v", decoded, deviceID)
		}
	})

	t.Run("DeviceID wrong size", func(t *testing.T) {
		wrongSizeID := make([]byte, 15) // Should be 16
		_, err := EncodeDeviceID(wrongSizeID)
		if err == nil {
			t.Error("Expected error for wrong size device ID")
		}
	})

	t.Run("UserToken", func(t *testing.T) {
		token := make([]byte, 32)
		for i := range token {
			token[i] = byte(i)
		}

		encoded, err := EncodeUserToken(token)
		if err != nil {
			t.Errorf("EncodeUserToken() error = %v", err)
		}

		decoded, err := DecodeUserToken(encoded)
		if err != nil {
			t.Errorf("DecodeUserToken() error = %v", err)
		}

		if !bytes.Equal(decoded, token) {
			t.Errorf("Round-trip failed: got %v, expected %v", decoded, token)
		}
	})

	t.Run("Ed25519Signature", func(t *testing.T) {
		signature := make([]byte, 64)
		for i := range signature {
			signature[i] = byte(i % 256)
		}

		encoded, err := EncodeEd25519Signature(signature)
		if err != nil {
			t.Errorf("EncodeEd25519Signature() error = %v", err)
		}

		decoded, err := DecodeEd25519Signature(encoded)
		if err != nil {
			t.Errorf("DecodeEd25519Signature() error = %v", err)
		}

		if !bytes.Equal(decoded, signature) {
			t.Errorf("Round-trip failed: got %v, expected %v", decoded, signature)
		}
	})

	t.Run("Ed25519PublicKey", func(t *testing.T) {
		publicKey := make([]byte, 32)
		for i := range publicKey {
			publicKey[i] = byte(i)
		}

		encoded, err := EncodeEd25519PublicKey(publicKey)
		if err != nil {
			t.Errorf("EncodeEd25519PublicKey() error = %v", err)
		}

		decoded, err := DecodeEd25519PublicKey(encoded)
		if err != nil {
			t.Errorf("DecodeEd25519PublicKey() error = %v", err)
		}

		if !bytes.Equal(decoded, publicKey) {
			t.Errorf("Round-trip failed: got %v, expected %v", decoded, publicKey)
		}
	})

	t.Run("X25519PublicKey", func(t *testing.T) {
		publicKey := make([]byte, 32)
		for i := range publicKey {
			publicKey[i] = byte(255 - i)
		}

		encoded, err := EncodeX25519PublicKey(publicKey)
		if err != nil {
			t.Errorf("EncodeX25519PublicKey() error = %v", err)
		}

		decoded, err := DecodeX25519PublicKey(encoded)
		if err != nil {
			t.Errorf("DecodeX25519PublicKey() error = %v", err)
		}

		if !bytes.Equal(decoded, publicKey) {
			t.Errorf("Round-trip failed: got %v, expected %v", decoded, publicKey)
		}
	})
}

func TestIsValidBase64URLChar(t *testing.T) {
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	invalidChars := "+/=!@#$%^&*()[]{}|\\:;\"'<>,.?`~"

	for _, char := range validChars {
		if !isValidBase64URLChar(char) {
			t.Errorf("Valid character %c was rejected", char)
		}
	}

	for _, char := range invalidChars {
		if isValidBase64URLChar(char) {
			t.Errorf("Invalid character %c was accepted", char)
		}
	}
}

// Helper functions

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
