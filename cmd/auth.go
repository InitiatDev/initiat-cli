package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/auth"
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

	if err := auth.AuthenticateUser(email, ""); err != nil {
		return err
	}

	fmt.Println("✅ Login successful! Registration token expires in 15 minutes.")
	fmt.Println("💡 Next: Register this device with 'initiat device register <name>'")

	return nil
}
