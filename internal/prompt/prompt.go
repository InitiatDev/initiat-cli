package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptEmail prompts the user for an email address
func PromptEmail() (string, error) {
	fmt.Print("Email: ")
	reader := bufio.NewReader(os.Stdin)
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}
	return email, nil
}

// PromptPassword prompts the user for a password (hidden input)
func PromptPassword() (string, error) {
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	password := string(passwordBytes)
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	return password, nil
}
