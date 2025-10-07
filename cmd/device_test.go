package cmd

import (
	"crypto/ed25519"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/types"
	"golang.org/x/crypto/curve25519"
)

func TestGenerateEd25519Keypair(t *testing.T) {
	publicKey, privateKey, err := generateEd25519Keypair()
	if err != nil {
		t.Fatalf("Failed to generate Ed25519 keypair: %v", err)
	}

	if len(publicKey) != ed25519.PublicKeySize {
		t.Errorf("Expected public key size %d, got %d", ed25519.PublicKeySize, len(publicKey))
	}

	if len(privateKey) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	message := []byte("test message")
	signature := ed25519.Sign(privateKey, message)

	if !ed25519.Verify(publicKey, message, signature) {
		t.Error("Generated keypair failed signature verification")
	}
}

func TestGenerateX25519Keypair(t *testing.T) {
	publicKey, privateKey, err := generateX25519Keypair()
	if err != nil {
		t.Fatalf("Failed to generate X25519 keypair: %v", err)
	}

	if len(publicKey) != crypto.X25519PrivateKeySize {
		t.Errorf("Expected public key size %d, got %d", crypto.X25519PrivateKeySize, len(publicKey))
	}

	if len(privateKey) != crypto.X25519PrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", crypto.X25519PrivateKeySize, len(privateKey))
	}

	expectedPublicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		t.Fatalf("Failed to compute expected public key: %v", err)
	}

	for i := 0; i < len(publicKey); i++ {
		if publicKey[i] != expectedPublicKey[i] {
			t.Error("Generated public key doesn't match expected value")
			break
		}
	}
}

func TestGenerateKeypairsUnique(t *testing.T) {
	pub1, priv1, err := generateEd25519Keypair()
	if err != nil {
		t.Fatalf("Failed to generate first Ed25519 keypair: %v", err)
	}

	pub2, priv2, err := generateEd25519Keypair()
	if err != nil {
		t.Fatalf("Failed to generate second Ed25519 keypair: %v", err)
	}

	if string(pub1) == string(pub2) {
		t.Error("Generated Ed25519 public keys are identical")
	}

	if string(priv1) == string(priv2) {
		t.Error("Generated Ed25519 private keys are identical")
	}

	xPub1, xPriv1, err := generateX25519Keypair()
	if err != nil {
		t.Fatalf("Failed to generate first X25519 keypair: %v", err)
	}

	xPub2, xPriv2, err := generateX25519Keypair()
	if err != nil {
		t.Fatalf("Failed to generate second X25519 keypair: %v", err)
	}

	if string(xPub1) == string(xPub2) {
		t.Error("Generated X25519 public keys are identical")
	}

	if string(xPriv1) == string(xPriv2) {
		t.Error("Generated X25519 private keys are identical")
	}
}

func createApproval(orgSlug, projectSlug string) types.DeviceApproval {
	approval := types.DeviceApproval{}
	approval.ProjectMembership.Project.Slug = projectSlug
	approval.ProjectMembership.Project.Organization.Slug = orgSlug
	return approval
}

func TestBuildProjectSlug(t *testing.T) {
	tests := []struct {
		name     string
		approval types.DeviceApproval
		expected string
	}{
		{
			name:     "normal organization and project",
			approval: createApproval("acme-corp", "production"),
			expected: "acme-corp/production",
		},
		{
			name:     "empty organization slug",
			approval: createApproval("", "default"),
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildProjectSlug(tt.approval)
			if result != tt.expected {
				t.Errorf("buildProjectSlug() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly", 7, "exactly"},
		{"very long string", 10, "very lo..."},
		{"", 5, ""},
		{"test", 0, ""},
		{"test", 1, "t"},
		{"test", 2, "te"},
		{"test", 3, "tes"},
		{"test", 4, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"2024-01-15T10:30:00Z",
			"Jan 15 10:30",
		},
		{
			"2024-12-25T23:59:59Z",
			"Dec 25 23:59",
		},
		{
			"invalid-time",
			"invalid-time",
		},
		{
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatTime(tt.input)
			if result != tt.expected {
				t.Errorf("formatTime(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
