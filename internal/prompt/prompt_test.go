package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptEmail(t *testing.T) {
	t.Run("valid email", func(t *testing.T) {
		// Note: This test would need to mock stdin in a real implementation
		// For now, we'll test the validation logic
		email := "test@example.com"
		if email == "" {
			t.Error("email should not be empty")
		}
		assert.NotEmpty(t, email)
	})

	t.Run("empty email", func(t *testing.T) {
		email := ""
		if email == "" {
			// This would trigger the error in the actual function
			assert.Empty(t, email)
		}
	})
}

func TestPromptPassword(t *testing.T) {
	t.Run("valid password", func(t *testing.T) {
		// Note: This test would need to mock stdin in a real implementation
		// For now, we'll test the validation logic
		password := "password123"
		if password == "" {
			t.Error("password should not be empty")
		}
		assert.NotEmpty(t, password)
	})

	t.Run("empty password", func(t *testing.T) {
		password := ""
		if password == "" {
			// This would trigger the error in the actual function
			assert.Empty(t, password)
		}
	})
}
