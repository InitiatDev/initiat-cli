package storage

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/config"
)

const (
	DefaultServiceName = "initiat-cli"
)

// TestKeyring is set by tests to inject an in-memory keyring so code that
// calls storage.New() (e.g. httputil.SignRequest) uses it. Must be nil in production.
var TestKeyring Keyring

type Storage struct {
	serviceName string
	kr          Keyring
}

func (s *Storage) keyring() Keyring {
	if s.kr != nil {
		return s.kr
	}
	if TestKeyring != nil {
		return TestKeyring
	}
	return systemKeyring{}
}

func generateServiceNameFromURL(apiURL string) string {
	parsed, err := url.Parse(apiURL)
	if err != nil {
		parsed = &url.URL{Host: strings.ReplaceAll(apiURL, "://", "-")}
	}

	host := parsed.Host
	if host == "" {
		host = strings.ReplaceAll(apiURL, "://", "-")
		host = strings.ReplaceAll(host, "/", "-")
	}

	hasher := sha256.New()
	hasher.Write([]byte(apiURL))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

	return fmt.Sprintf("initiat-cli-%s-%s", host, hash)
}

func New() *Storage {
	cfg := config.Get()

	var serviceName string
	if cfg.ServiceName != "" && cfg.ServiceName != DefaultServiceName {
		serviceName = cfg.ServiceName
	} else {
		apiURL := config.GetAPIBaseURL()
		serviceName = generateServiceNameFromURL(apiURL)
	}

	s := &Storage{serviceName: serviceName}
	if TestKeyring != nil {
		s.kr = TestKeyring
	}
	return s
}

func NewWithServiceName(serviceName string) *Storage {
	s := &Storage{serviceName: serviceName}
	if TestKeyring != nil {
		s.kr = TestKeyring
	}
	return s
}

func NewWithKeyring(serviceName string, kr Keyring) *Storage {
	return &Storage{
		serviceName: serviceName,
		kr:          kr,
	}
}

func (s *Storage) StoreToken(token string) error {
	return s.keyring().Set(s.serviceName, "registration-token", token)
}

func (s *Storage) GetToken() (string, error) {
	token, err := s.keyring().Get(s.serviceName, "registration-token")
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token, nil
}

func (s *Storage) DeleteToken() error {
	return s.keyring().Delete(s.serviceName, "registration-token")
}

func (s *Storage) StoreDeviceID(deviceID string) error {
	return s.keyring().Set(s.serviceName, "device-id", deviceID)
}

func (s *Storage) GetDeviceID() (string, error) {
	deviceID, err := s.keyring().Get(s.serviceName, "device-id")
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return deviceID, nil
}

func (s *Storage) DeleteDeviceID() error {
	return s.keyring().Delete(s.serviceName, "device-id")
}

func (s *Storage) StoreDeviceName(deviceName string) error {
	return s.keyring().Set(s.serviceName, "device-name", deviceName)
}

func (s *Storage) GetDeviceName() (string, error) {
	deviceName, err := s.keyring().Get(s.serviceName, "device-name")
	if err != nil {
		return "", fmt.Errorf("failed to get device name: %w", err)
	}
	return deviceName, nil
}

func (s *Storage) DeleteDeviceName() error {
	return s.keyring().Delete(s.serviceName, "device-name")
}

func (s *Storage) HasDeviceName() bool {
	_, err := s.GetDeviceName()
	return err == nil
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
	return s.keyring().Set(s.serviceName, "signing-private-key", string(privateKey))
}

func (s *Storage) GetSigningPrivateKey() (ed25519.PrivateKey, error) {
	keyStr, err := s.keyring().Get(s.serviceName, "signing-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get signing private key: %w", err)
	}
	return ed25519.PrivateKey(keyStr), nil
}

func (s *Storage) DeleteSigningPrivateKey() error {
	return s.keyring().Delete(s.serviceName, "signing-private-key")
}

func (s *Storage) StoreEncryptionPrivateKey(privateKey []byte) error {
	return s.keyring().Set(s.serviceName, "encryption-private-key", string(privateKey))
}

func (s *Storage) GetEncryptionPrivateKey() ([]byte, error) {
	keyStr, err := s.keyring().Get(s.serviceName, "encryption-private-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption private key: %w", err)
	}
	return []byte(keyStr), nil
}

func (s *Storage) DeleteEncryptionPrivateKey() error {
	return s.keyring().Delete(s.serviceName, "encryption-private-key")
}

func (s *Storage) HasSigningPrivateKey() bool {
	_, err := s.GetSigningPrivateKey()
	return err == nil
}

func (s *Storage) HasEncryptionPrivateKey() bool {
	_, err := s.GetEncryptionPrivateKey()
	return err == nil
}

func (s *Storage) ClearDeviceCredentials() error {
	var errors []error

	if err := s.DeleteDeviceID(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete device ID: %w", err))
	}

	if err := s.DeleteDeviceName(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete device name: %w", err))
	}

	if err := s.DeleteSigningPrivateKey(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete signing private key: %w", err))
	}

	if err := s.DeleteEncryptionPrivateKey(); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete encryption private key: %w", err))
	}

	_ = s.DeleteToken()

	if len(errors) > 0 {
		return fmt.Errorf("errors clearing device credentials: %v", errors)
	}

	return nil
}
