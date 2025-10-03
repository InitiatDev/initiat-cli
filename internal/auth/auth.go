package auth

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
	"github.com/InitiatDev/initiat-cli/internal/storage"
)

// AuthenticateUser handles the complete authentication flow
func AuthenticateUser(email, password string) error {
	if email == "" {
		var err error
		email, err = prompt.PromptEmail()
		if err != nil {
			return err
		}
	}

	if password == "" {
		var err error
		password, err = prompt.PromptPassword()
		if err != nil {
			return err
		}
	}

	fmt.Println("🔐 Authenticating...")

	apiClient := client.New()
	loginResp, err := apiClient.Login(email, password)
	if err != nil {
		return fmt.Errorf("❌ Authentication failed: %w", err)
	}

	store := storage.New()
	if err := store.StoreToken(loginResp.Token); err != nil {
		return fmt.Errorf("❌ Failed to store authentication token: %w", err)
	}

	fmt.Printf("✅ Authenticated as %s %s\n", loginResp.User.Name, loginResp.User.Surname)
	fmt.Println()

	return nil
}

// EnsureAuthenticated checks if user is authenticated, prompts if not
func EnsureAuthenticated() error {
	store := storage.New()

	if store.HasToken() {
		fmt.Println("ℹ️  Found existing authentication token")
		return nil
	}

	fmt.Println("🔐 Authentication required for device registration")
	fmt.Println()

	return AuthenticateUser("", "")
}
