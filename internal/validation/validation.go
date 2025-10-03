package validation

import (
	"fmt"
	"strings"
)

// ValidateDeviceName validates a device name
func ValidateDeviceName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("device name cannot be empty")
	}
	return nil
}

// ValidateSecretKey validates a secret key
func ValidateSecretKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return fmt.Errorf("secret key cannot be empty")
	}
	return nil
}

// ValidateSecretValue validates a secret value
func ValidateSecretValue(value string) error {
	if value == "" {
		return fmt.Errorf("secret value cannot be empty")
	}
	return nil
}
