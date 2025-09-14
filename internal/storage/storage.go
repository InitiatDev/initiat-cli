package storage

import (
	"crypto/ed25519"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "initflow-cli"
)

// Storage handles secure storage of sensitive data
type Storage struct{}

// New creates a new storage instance
func New() *Storage {
	return &Storage{}
}

// StoreToken stores the registration token securely
func (s *Storage) StoreToken(token string) error {
	return keyring.Set(serviceName, "registration-token", token)
}

// GetToken retrieves the registration token
func (s *Storage) GetToken() (string, error) {
	token, err := keyring.Get(serviceName, "registration-token")
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token, nil
}

// DeleteToken removes the registration token
func (s *Storage) DeleteToken() error {
	return keyring.Delete(serviceName, "registration-token")
}

// StoreDeviceID stores the device ID
func (s *Storage) StoreDeviceID(deviceID string) error {
	return keyring.Set(serviceName, "device-id", deviceID)
}

// GetDeviceID retrieves the device ID
func (s *Storage) GetDeviceID() (string, error) {
	deviceID, err := keyring.Get(serviceName, "device-id")
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return deviceID, nil
}

// DeleteDeviceID removes the device ID
func (s *Storage) DeleteDeviceID() error {
	return keyring.Delete(serviceName, "device-id")
}

// HasToken checks if a registration token exists
func (s *Storage) HasToken() bool {
	_, err := s.GetToken()
	return err == nil
}

// HasDeviceID checks if a device ID exists
func (s *Storage) HasDeviceID() bool {
	_, err := s.GetDeviceID()
	return err == nil
}

// StoreSigningPrivateKey stores the Ed25519 signing private key
func (s *Storage) StoreSigningPrivateKey(privateKey ed25519.PrivateKey) error {
	return keyring.Set(serviceName, "signing-private-key", string(privateKey))
}

// GetSigningPrivateKey retrieves the Ed25519 signing private key
func (s *Storage) GetSigningPrivateKey() (ed25519.PrivateKey, error) {
	keyStr, err := keyring.Get(serviceName, "signing-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get signing private key: %w", err)
	}
	return ed25519.PrivateKey(keyStr), nil
}

// DeleteSigningPrivateKey removes the Ed25519 signing private key
func (s *Storage) DeleteSigningPrivateKey() error {
	return keyring.Delete(serviceName, "signing-private-key")
}

// StoreEncryptionPrivateKey stores the X25519 encryption private key
func (s *Storage) StoreEncryptionPrivateKey(privateKey []byte) error {
	return keyring.Set(serviceName, "encryption-private-key", string(privateKey))
}

// GetEncryptionPrivateKey retrieves the X25519 encryption private key
func (s *Storage) GetEncryptionPrivateKey() ([]byte, error) {
	keyStr, err := keyring.Get(serviceName, "encryption-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption private key: %w", err)
	}
	return []byte(keyStr), nil
}

// DeleteEncryptionPrivateKey removes the X25519 encryption private key
func (s *Storage) DeleteEncryptionPrivateKey() error {
	return keyring.Delete(serviceName, "encryption-private-key")
}

// HasSigningPrivateKey checks if a signing private key exists
func (s *Storage) HasSigningPrivateKey() bool {
	_, err := s.GetSigningPrivateKey()
	return err == nil
}

// HasEncryptionPrivateKey checks if an encryption private key exists
func (s *Storage) HasEncryptionPrivateKey() bool {
	_, err := s.GetEncryptionPrivateKey()
	return err == nil
}
