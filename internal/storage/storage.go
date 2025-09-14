package storage

import (
	"crypto/ed25519"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "initflow-cli"
)

type Storage struct{}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) StoreToken(token string) error {
	return keyring.Set(serviceName, "registration-token", token)
}

func (s *Storage) GetToken() (string, error) {
	token, err := keyring.Get(serviceName, "registration-token")
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token, nil
}

func (s *Storage) DeleteToken() error {
	return keyring.Delete(serviceName, "registration-token")
}

func (s *Storage) StoreDeviceID(deviceID string) error {
	return keyring.Set(serviceName, "device-id", deviceID)
}

func (s *Storage) GetDeviceID() (string, error) {
	deviceID, err := keyring.Get(serviceName, "device-id")
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return deviceID, nil
}

func (s *Storage) DeleteDeviceID() error {
	return keyring.Delete(serviceName, "device-id")
}

func (s *Storage) HasToken() bool {
	_, err := s.GetToken()
	return err == nil
}

func (s *Storage) HasDeviceID() bool {
	_, err := s.GetDeviceID()
	return err == nil
}

func (s *Storage) StoreSigningPrivateKey(privateKey ed25519.PrivateKey) error {
	return keyring.Set(serviceName, "signing-private-key", string(privateKey))
}

func (s *Storage) GetSigningPrivateKey() (ed25519.PrivateKey, error) {
	keyStr, err := keyring.Get(serviceName, "signing-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get signing private key: %w", err)
	}
	return ed25519.PrivateKey(keyStr), nil
}

func (s *Storage) DeleteSigningPrivateKey() error {
	return keyring.Delete(serviceName, "signing-private-key")
}

func (s *Storage) StoreEncryptionPrivateKey(privateKey []byte) error {
	return keyring.Set(serviceName, "encryption-private-key", string(privateKey))
}

func (s *Storage) GetEncryptionPrivateKey() ([]byte, error) {
	keyStr, err := keyring.Get(serviceName, "encryption-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption private key: %w", err)
	}
	return []byte(keyStr), nil
}

func (s *Storage) DeleteEncryptionPrivateKey() error {
	return keyring.Delete(serviceName, "encryption-private-key")
}

func (s *Storage) HasSigningPrivateKey() bool {
	_, err := s.GetSigningPrivateKey()
	return err == nil
}

func (s *Storage) HasEncryptionPrivateKey() bool {
	_, err := s.GetEncryptionPrivateKey()
	return err == nil
}

func (s *Storage) StoreWorkspaceKey(workspaceSlug string, key []byte) error {
	keyName := fmt.Sprintf("workspace-key-%s", workspaceSlug)
	return keyring.Set(serviceName, keyName, string(key))
}

func (s *Storage) GetWorkspaceKey(workspaceSlug string) ([]byte, error) {
	keyName := fmt.Sprintf("workspace-key-%s", workspaceSlug)
	keyStr, err := keyring.Get(serviceName, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace key for %s: %w", workspaceSlug, err)
	}
	return []byte(keyStr), nil
}

func (s *Storage) DeleteWorkspaceKey(workspaceSlug string) error {
	keyName := fmt.Sprintf("workspace-key-%s", workspaceSlug)
	return keyring.Delete(serviceName, keyName)
}

func (s *Storage) HasWorkspaceKey(workspaceSlug string) bool {
	_, err := s.GetWorkspaceKey(workspaceSlug)
	return err == nil
}
