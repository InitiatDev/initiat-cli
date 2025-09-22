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

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/config"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device management commands",
	Long:  "Manage devices registered with Initiat",
}

var registerDeviceCmd = &cobra.Command{
	Use:   "register <device-name>",
	Short: "Register this device with Initiat",
	Long:  "Register this device with Initiat to enable secure secret access",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegisterDevice,
}

var unregisterDeviceCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Clear local device credentials",
	Long: "Remove all device credentials stored locally in the system keychain. " +
		"Use this when you want to register a fresh device or clean up after deleting a device from the server.",
	RunE: runUnregisterDevice,
}

var clearTokenCmd = &cobra.Command{
	Use:   "clear-token",
	Short: "Clear stored authentication token",
	Long: "Remove the stored authentication token. " +
		"Use this if you're getting 'Invalid or expired registration token' errors.",
	RunE: runClearToken,
}

func init() {
	rootCmd.AddCommand(deviceCmd)
	deviceCmd.AddCommand(registerDeviceCmd)
	deviceCmd.AddCommand(unregisterDeviceCmd)
	deviceCmd.AddCommand(clearTokenCmd)
}

func ensureAuthenticated() error {
	storage := storage.New()

	if storage.HasToken() {
		fmt.Println("ℹ️  Found existing authentication token")
		return nil
	}

	fmt.Println("🔐 Authentication required for device registration")
	fmt.Println()

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

	if err := storage.StoreToken(loginResp.Token); err != nil {
		return fmt.Errorf("❌ Failed to store authentication token: %w", err)
	}

	fmt.Printf("✅ Authenticated as %s %s\n", loginResp.User.Name, loginResp.User.Surname)
	fmt.Println()

	return nil
}

func generateEd25519Keypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 keypair: %w", err)
	}
	return publicKey, privateKey, nil
}

const x25519KeySize = 32

func generateX25519Keypair() ([]byte, []byte, error) {
	privateKey := make([]byte, x25519KeySize)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate X25519 private key: %w", err)
	}

	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate X25519 public key: %w", err)
	}

	return publicKey, privateKey, nil
}

func checkExistingDevice(storage *storage.Storage) error {
	if !storage.HasDeviceID() {
		return nil
	}

	deviceID, _ := storage.GetDeviceID()
	fmt.Printf("⚠️  Device already registered with ID: %s\n", deviceID)
	fmt.Println()
	fmt.Println("If you deleted this device from the server, you can:")
	fmt.Println("• Clear local credentials: initiat device unregister")
	fmt.Println("• Then register again: initiat device register <name>")
	fmt.Println()
	fmt.Println("Or use 'initiat device list' to view registered devices")
	return fmt.Errorf("device already registered")
}

func generateKeypairs() (ed25519.PublicKey, ed25519.PrivateKey, []byte, []byte, error) {
	fmt.Println("🔑 Generating Ed25519 signing keypair...")
	signingPublicKey, signingPrivateKey, err := generateEd25519Keypair()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate signing keypair: %w", err)
	}

	fmt.Println("🔒 Generating X25519 encryption keypair...")
	encryptionPublicKey, encryptionPrivateKey, err := generateX25519Keypair()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate encryption keypair: %w", err)
	}

	return signingPublicKey, signingPrivateKey, encryptionPublicKey, encryptionPrivateKey, nil
}

func performDeviceRegistration(
	deviceName string,
	signingPublicKey ed25519.PublicKey,
	encryptionPublicKey []byte,
	storage *storage.Storage,
) (*client.DeviceRegistrationResponse, error) {
	fmt.Println("📡 Registering device with server...")
	apiClient := client.New()

	// Debug: show current config
	cfg := config.Get()
	fmt.Printf("🔍 Debug: API URL: %s\n", cfg.APIBaseURL)

	token, err := storage.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication token: %w", err)
	}

	fmt.Printf("🔍 Debug: Using API client with token length: %d\n", len(token))
	fmt.Printf("🔍 Debug: Ed25519 public key size: %d bytes\n", len(signingPublicKey))
	fmt.Printf("🔍 Debug: X25519 public key size: %d bytes\n", len(encryptionPublicKey))

	deviceResp, err := apiClient.RegisterDevice(token, deviceName, signingPublicKey, encryptionPublicKey)
	if err != nil {
		fmt.Printf("🔍 Debug: Registration error details: %v\n", err)
		return nil, fmt.Errorf("❌ Device registration failed: %w", err)
	}

	return deviceResp, nil
}

func storeDeviceCredentials(
	storage *storage.Storage,
	signingPrivateKey ed25519.PrivateKey,
	encryptionPrivateKey []byte,
	deviceID string,
) error {
	fmt.Println("🔐 Storing keys securely in system keychain...")

	if err := storage.StoreSigningPrivateKey(signingPrivateKey); err != nil {
		return fmt.Errorf("failed to store signing private key: %w", err)
	}

	if err := storage.StoreEncryptionPrivateKey(encryptionPrivateKey); err != nil {
		return fmt.Errorf("failed to store encryption private key: %w", err)
	}

	if err := storage.StoreDeviceID(deviceID); err != nil {
		return fmt.Errorf("failed to store device ID: %w", err)
	}

	return nil
}

func runRegisterDevice(cmd *cobra.Command, args []string) error {
	deviceName := strings.TrimSpace(args[0])
	if deviceName == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	if err := ensureAuthenticated(); err != nil {
		return err
	}

	storage := storage.New()

	if err := checkExistingDevice(storage); err != nil {
		return nil // Not a real error, just early return
	}

	fmt.Printf("🔑 Registering device: %s\n", deviceName)

	signingPublicKey, signingPrivateKey, encryptionPublicKey, encryptionPrivateKey, err := generateKeypairs()
	if err != nil {
		return err
	}

	deviceResp, err := performDeviceRegistration(deviceName, signingPublicKey, encryptionPublicKey, storage)
	if err != nil {
		return err
	}

	err = storeDeviceCredentials(storage, signingPrivateKey, encryptionPrivateKey, deviceResp.Device.DeviceID)
	if err != nil {
		return err
	}

	_ = storage.DeleteToken()
	fmt.Println("✅ Device registered successfully!")
	fmt.Println()
	fmt.Printf("Device ID: %s\n", deviceResp.Device.DeviceID)
	fmt.Printf("Device Name: %s\n", deviceResp.Device.Name)
	fmt.Printf("Created: %s\n", deviceResp.Device.CreatedAt)
	fmt.Println()
	fmt.Println("🔐 Keys stored securely in system keychain")
	fmt.Println("💡 Next: Initialize workspace keys with 'initiat workspace list'")

	return nil
}

func runUnregisterDevice(cmd *cobra.Command, args []string) error {
	storage := storage.New()

	// Check if there are any device credentials to clear
	if !storage.HasDeviceID() && !storage.HasSigningPrivateKey() && !storage.HasEncryptionPrivateKey() {
		fmt.Println("ℹ️  No device credentials found in local storage")
		return nil
	}

	fmt.Println("🔐 Clearing local device credentials...")

	err := storage.ClearDeviceCredentials()
	if err != nil {
		return fmt.Errorf("❌ Failed to clear device credentials: %w", err)
	}

	fmt.Println("✅ Device credentials cleared successfully!")
	fmt.Println()
	fmt.Println("💡 You can now register a new device with 'initiat device register <name>'")

	return nil
}

func runClearToken(cmd *cobra.Command, args []string) error {
	storage := storage.New()

	if !storage.HasToken() {
		fmt.Println("ℹ️  No authentication token found in local storage")
		return nil
	}

	fmt.Println("🔐 Clearing authentication token...")

	err := storage.DeleteToken()
	if err != nil {
		return fmt.Errorf("❌ Failed to clear authentication token: %w", err)
	}

	fmt.Println("✅ Authentication token cleared successfully!")
	fmt.Println("💡 You will need to authenticate again for device registration")

	return nil
}
