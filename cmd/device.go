package cmd

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/term"

	"github.com/DylanBlakemore/initflow-cli/internal/client"
	"github.com/DylanBlakemore/initflow-cli/internal/storage"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device management commands",
	Long:  "Manage devices registered with InitFlow",
}

var registerDeviceCmd = &cobra.Command{
	Use:   "register <device-name>",
	Short: "Register this device with InitFlow",
	Long:  "Register this device with InitFlow to enable secure secret access",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegisterDevice,
}

func init() {
	rootCmd.AddCommand(deviceCmd)
	deviceCmd.AddCommand(registerDeviceCmd)
}

// ensureAuthenticated checks if user is authenticated and prompts for login if not
func ensureAuthenticated() error {
	storage := storage.New()

	// Check if we have a valid token
	if storage.HasToken() {
		return nil
	}

	fmt.Println("🔐 Authentication required for device registration")
	fmt.Println()

	// Prompt for email
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Prompt for password
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	password := string(passwordBytes)
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Show authentication progress
	fmt.Println("🔐 Authenticating...")

	// Create client and attempt login
	apiClient := client.New()
	loginResp, err := apiClient.Login(email, password)
	if err != nil {
		return fmt.Errorf("❌ Authentication failed: %w", err)
	}

	// Store the registration token securely
	if err := storage.StoreToken(loginResp.Token); err != nil {
		return fmt.Errorf("❌ Failed to store authentication token: %w", err)
	}

	fmt.Printf("✅ Authenticated as %s %s\n", loginResp.User.Name, loginResp.User.Surname)
	fmt.Println()

	return nil
}

// generateEd25519Keypair generates an Ed25519 signing keypair
func generateEd25519Keypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 keypair: %w", err)
	}
	return publicKey, privateKey, nil
}

// generateX25519Keypair generates an X25519 encryption keypair
func generateX25519Keypair() ([]byte, []byte, error) {
	// Generate private key (32 random bytes)
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate X25519 private key: %w", err)
	}

	// Generate public key from private key
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate X25519 public key: %w", err)
	}

	return publicKey, privateKey, nil
}

func runRegisterDevice(cmd *cobra.Command, args []string) error {
	deviceName := strings.TrimSpace(args[0])
	if deviceName == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	// Ensure user is authenticated
	if err := ensureAuthenticated(); err != nil {
		return err
	}

	// Check if device is already registered
	storage := storage.New()
	if storage.HasDeviceID() {
		deviceID, _ := storage.GetDeviceID()
		fmt.Printf("⚠️  Device already registered with ID: %s\n", deviceID)
		fmt.Println("💡 Use 'initflow device list' to view registered devices")
		return nil
	}

	fmt.Printf("🔑 Registering device: %s\n", deviceName)

	// Generate Ed25519 signing keypair
	fmt.Println("🔑 Generating Ed25519 signing keypair...")
	signingPublicKey, signingPrivateKey, err := generateEd25519Keypair()
	if err != nil {
		return fmt.Errorf("failed to generate signing keypair: %w", err)
	}

	// Generate X25519 encryption keypair
	fmt.Println("🔒 Generating X25519 encryption keypair...")
	encryptionPublicKey, encryptionPrivateKey, err := generateX25519Keypair()
	if err != nil {
		return fmt.Errorf("failed to generate encryption keypair: %w", err)
	}

	// Register device with server
	fmt.Println("📡 Registering device with server...")
	apiClient := client.New()

	token, err := storage.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}

	deviceResp, err := apiClient.RegisterDevice(token, deviceName, signingPublicKey, encryptionPublicKey)
	if err != nil {
		return fmt.Errorf("❌ Device registration failed: %w", err)
	}

	// Store private keys and device ID securely
	fmt.Println("🔐 Storing keys securely in system keychain...")

	if err := storage.StoreSigningPrivateKey(signingPrivateKey); err != nil {
		return fmt.Errorf("failed to store signing private key: %w", err)
	}

	if err := storage.StoreEncryptionPrivateKey(encryptionPrivateKey); err != nil {
		return fmt.Errorf("failed to store encryption private key: %w", err)
	}

	if err := storage.StoreDeviceID(deviceResp.Device.DeviceID); err != nil {
		return fmt.Errorf("failed to store device ID: %w", err)
	}

	// Clean up registration token (it's no longer needed)
	_ = storage.DeleteToken()

	// Success message
	fmt.Println("✅ Device registered successfully!")
	fmt.Println()
	fmt.Printf("Device ID: %s\n", deviceResp.Device.DeviceID)
	fmt.Printf("Device Name: %s\n", deviceResp.Device.Name)
	fmt.Printf("Created: %s\n", deviceResp.Device.CreatedAt)
	fmt.Println()
	fmt.Println("🔐 Keys stored securely in system keychain")
	fmt.Println("💡 Next: Initialize workspace keys with 'initflow workspace list'")

	return nil
}
