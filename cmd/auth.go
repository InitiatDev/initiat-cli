package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage authentication with Initiat",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Initiat",
	Long: `Authenticate with Initiat using your email and password.

If --email is not provided, you will be prompted for it.
Password is always prompted securely.

Examples:
  initiat auth login --email user@example.com
  initiat auth login -e user@example.com
  initiat auth login  # Will prompt for email`,
	RunE: runLogin,
}

var (
	loginEmail string
)

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Email address for login")
}

func runLogin(cmd *cobra.Command, args []string) error {
	email := strings.TrimSpace(loginEmail)

	if email == "" {
		fmt.Print("Email: ")
		_, err := fmt.Scanln(&email)
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(email)
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	password := string(passwordBytes)
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Println("🔐 Authenticating...")

	apiClient := client.New()
	loginResp, err := apiClient.Login(email, password)
	if err != nil {
		return fmt.Errorf("❌ Authentication failed: %w", err)
	}

	storage := storage.New()
	if err := storage.StoreToken(loginResp.Token); err != nil {
		return fmt.Errorf("❌ Failed to store authentication token: %w", err)
	}
	fmt.Println("✅ Login successful! Registration token expires in 15 minutes.")
	fmt.Printf("👋 Welcome, %s %s!\n", loginResp.User.Name, loginResp.User.Surname)
	fmt.Println("💡 Next: Register this device with 'initiat device register <name>'")

	return nil
}
