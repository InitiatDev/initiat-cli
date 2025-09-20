package storage

import (
	"crypto/ed25519"
	"fmt"

	"github.com/DylanBlakemore/initflow-cli/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	DefaultServiceName = "initflow-cli"
)

type Storage struct {
	serviceName string
}

func New() *Storage {
	serviceName := DefaultServiceName
	cfg := config.Get()
	if cfg.ServiceName != "" {
		serviceName = cfg.ServiceName
	}
	return &Storage{
		serviceName: serviceName,
	}
}

func NewWithServiceName(serviceName string) *Storage {
	return &Storage{
		serviceName: serviceName,
	}
}

func (s *Storage) StoreToken(token string) error {
	return keyring.Set(s.serviceName, "registration-token", token)
}

func (s *Storage) GetToken() (string, error) {
	token, err := keyring.Get(s.serviceName, "registration-token")
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token, nil
}

func (s *Storage) DeleteToken() error {
	return keyring.Delete(s.serviceName, "registration-token")
}

func (s *Storage) StoreDeviceID(deviceID string) error {
	return keyring.Set(s.serviceName, "device-id", deviceID)
}

func (s *Storage) GetDeviceID() (string, error) {
	deviceID, err := keyring.Get(s.serviceName, "device-id")
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return deviceID, nil
}

func (s *Storage) DeleteDeviceID() error {
	return keyring.Delete(s.serviceName, "device-id")
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
	return keyring.Set(s.serviceName, "signing-private-key", string(privateKey))
}

func (s *Storage) GetSigningPrivateKey() (ed25519.PrivateKey, error) {
	keyStr, err := keyring.Get(s.serviceName, "signing-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get signing private key: %w", err)
	}
	return ed25519.PrivateKey(keyStr), nil
}

func (s *Storage) DeleteSigningPrivateKey() error {
	return keyring.Delete(s.serviceName, "signing-private-key")
}

func (s *Storage) StoreEncryptionPrivateKey(privateKey []byte) error {
	return keyring.Set(s.serviceName, "encryption-private-key", string(privateKey))
}

func (s *Storage) GetEncryptionPrivateKey() ([]byte, error) {
	keyStr, err := keyring.Get(s.serviceName, "encryption-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption private key: %w", err)
	}
	return []byte(keyStr), nil
}

func (s *Storage) DeleteEncryptionPrivateKey() error {
	return keyring.Delete(s.serviceName, "encryption-private-key")
}

func (s *Storage) HasSigningPrivateKey() bool {
	_, err := s.GetSigningPrivateKey()
	return err == nil
}

func (s *Storage) HasEncryptionPrivateKey() bool {
	_, err := s.GetEncryptionPrivateKey()
	return err == nil
}

// StoreWorkspaceKey stores a workspace key using composite slug format
func (s *Storage) StoreWorkspaceKey(compositeSlug string, key []byte) error {
	keyName := fmt.Sprintf("workspace-key-%s", compositeSlug)
	return keyring.Set(s.serviceName, keyName, string(key))
}

// GetWorkspaceKey retrieves a workspace key using composite slug format
func (s *Storage) GetWorkspaceKey(compositeSlug string) ([]byte, error) {
	keyName := fmt.Sprintf("workspace-key-%s", compositeSlug)
	keyStr, err := keyring.Get(s.serviceName, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace key for %s: %w", compositeSlug, err)
	}
	return []byte(keyStr), nil
}

// DeleteWorkspaceKey deletes a workspace key using composite slug format
func (s *Storage) DeleteWorkspaceKey(compositeSlug string) error {
	keyName := fmt.Sprintf("workspace-key-%s", compositeSlug)
	return keyring.Delete(s.serviceName, keyName)
}

// HasWorkspaceKey checks if a workspace key exists using composite slug format
func (s *Storage) HasWorkspaceKey(compositeSlug string) bool {
	_, err := s.GetWorkspaceKey(compositeSlug)
	return err == nil
}

// ClearDeviceCredentials removes all device-related credentials from local storage
func (s *Storage) ClearDeviceCredentials() error {
	var errors []error

	if err := s.DeleteDeviceID(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete device ID: %w", err))
	}

	if err := s.DeleteSigningPrivateKey(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete signing private key: %w", err))
	}

	if err := s.DeleteEncryptionPrivateKey(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete encryption private key: %w", err))
	}

	// Also clean up any leftover registration token
	_ = s.DeleteToken() // Ignore error as token might not exist

	if len(errors) > 0 {
		return fmt.Errorf("errors clearing device credentials: %v", errors)
	}

	return nil
}
