package storage

import (
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
