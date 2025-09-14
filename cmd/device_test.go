package cmd

import (
	"crypto/ed25519"
	"testing"

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

	if len(publicKey) != x25519KeySize {
		t.Errorf("Expected public key size %d, got %d", x25519KeySize, len(publicKey))
	}

	if len(privateKey) != x25519KeySize {
		t.Errorf("Expected private key size %d, got %d", x25519KeySize, len(privateKey))
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
