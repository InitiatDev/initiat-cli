package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/DylanBlakemore/initflow-cli/internal/client"
	"github.com/DylanBlakemore/initflow-cli/internal/storage"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage authentication with InitFlow",
}

var loginCmd = &cobra.Command{
	Use:   "login <email>",
	Short: "Login to InitFlow",
	Long:  "Authenticate with InitFlow using your email and password",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogin,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	email := strings.TrimSpace(args[0])
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
	fmt.Println("💡 Next: Register this device with 'initflow device register <name>'")

	return nil
}
