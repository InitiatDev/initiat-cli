package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDeviceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "valid device name",
			input:    "my-laptop",
			expected: nil,
		},
		{
			name:     "valid device name with spaces",
			input:    "  my-laptop  ",
			expected: nil,
		},
		{
			name:     "empty device name",
			input:    "",
			expected: fmt.Errorf("device name cannot be empty"),
		},
		{
			name:     "whitespace only device name",
			input:    "   ",
			expected: fmt.Errorf("device name cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceName(tt.input)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expected.Error(), err.Error())
			}
		})
	}
}

func TestValidateSecretKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "valid secret key",
			input:    "API_KEY",
			expected: nil,
		},
		{
			name:     "valid secret key with spaces",
			input:    "  API_KEY  ",
			expected: nil,
		},
		{
			name:     "empty secret key",
			input:    "",
			expected: fmt.Errorf("secret key cannot be empty"),
		},
		{
			name:     "whitespace only secret key",
			input:    "   ",
			expected: fmt.Errorf("secret key cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretKey(tt.input)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expected.Error(), err.Error())
			}
		})
	}
}

func TestValidateSecretValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "valid secret value",
			input:    "sk-1234567890",
			expected: nil,
		},
		{
			name:     "empty secret value",
			input:    "",
			expected: fmt.Errorf("secret value cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretValue(tt.input)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expected.Error(), err.Error())
			}
		})
	}
}
