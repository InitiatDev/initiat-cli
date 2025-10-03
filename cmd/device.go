package cmd

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/term"

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/config"
	"github.com/DylanBlakemore/initiat-cli/internal/encoding"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
	"github.com/DylanBlakemore/initiat-cli/internal/table"
	"github.com/DylanBlakemore/initiat-cli/internal/types"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device management commands",
	Long:  "Manage devices registered with Initiat",
}

var registerDeviceCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this device with Initiat",
	Long: `Register this device with Initiat to enable secure secret access.

Examples:
  initiat device register --name "my-laptop"
  initiat device register -n "work-macbook"`,
	RunE: runRegisterDevice,
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

var approvalsCmd = &cobra.Command{
	Use:   "approvals",
	Short: "List pending device approvals",
	Long:  "List all pending device approvals for workspaces where you are an admin.",
	RunE:  runListApprovals,
}

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a device for workspace access",
	Long:  "Approve a specific device or all pending devices for workspace access.",
	RunE:  runApproveDevice,
}

var rejectCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject a device for workspace access",
	Long:  "Reject a specific device or all pending devices for workspace access.",
	RunE:  runRejectDevice,
}

var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "Show device approval details",
	Long:  "Show detailed information about a specific device approval.",
	RunE:  runShowApproval,
}

const (
	statusPending       = "pending"
	maxDisplayLength    = 15
	maxKeyDisplayLength = 20
	minTruncateLength   = 3
)

var (
	deviceName string
	approveAll bool
	rejectAll  bool
	approvalID string
)

func init() {
	rootCmd.AddCommand(deviceCmd)
	deviceCmd.AddCommand(registerDeviceCmd)
	deviceCmd.AddCommand(unregisterDeviceCmd)
	deviceCmd.AddCommand(clearTokenCmd)
	deviceCmd.AddCommand(approvalsCmd)
	deviceCmd.AddCommand(approveCmd)
	deviceCmd.AddCommand(rejectCmd)
	deviceCmd.AddCommand(approvalCmd)

	registerDeviceCmd.Flags().StringVarP(&deviceName, "name", "n", "", "Name for this device (required)")
	_ = registerDeviceCmd.MarkFlagRequired("name")

	approveCmd.Flags().BoolVar(&approveAll, "all", false, "Approve all pending devices")
	approveCmd.Flags().StringVar(&approvalID, "id", "", "Device approval ID to approve")

	rejectCmd.Flags().BoolVar(&rejectAll, "all", false, "Reject all pending devices")
	rejectCmd.Flags().StringVar(&approvalID, "id", "", "Device approval ID to reject")

	approvalCmd.Flags().StringVar(&approvalID, "id", "", "Device approval ID to show (required)")
	_ = approvalCmd.MarkFlagRequired("id")
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
) (*types.DeviceRegistrationResponse, error) {
	fmt.Println("📡 Registering device with server...")
	apiClient := client.New()

	cfg := config.Get()
	fmt.Printf("🔍 Debug: API URL: %s\n", cfg.API.BaseURL)

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
	name := strings.TrimSpace(deviceName)
	if name == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	if err := ensureAuthenticated(); err != nil {
		return err
	}

	storage := storage.New()

	if err := checkExistingDevice(storage); err != nil {
		return nil
	}

	fmt.Printf("🔑 Registering device: %s\n", name)

	signingPublicKey, signingPrivateKey, encryptionPublicKey, encryptionPrivateKey, err := generateKeypairs()
	if err != nil {
		return err
	}

	deviceResp, err := performDeviceRegistration(name, signingPublicKey, encryptionPublicKey, storage)
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

func runListApprovals(cmd *cobra.Command, args []string) error {
	apiClient := client.New()

	approvals, err := apiClient.ListDeviceApprovals()
	if err != nil {
		return fmt.Errorf("❌ Failed to list device approvals: %w", err)
	}

	if len(approvals) == 0 {
		fmt.Println("📋 No pending device approvals found")
		return nil
	}

	fmt.Printf("📋 Pending Device Approvals (%d)\n\n", len(approvals))

	t := table.New()
	t.SetHeaders("ID", "User", "Device", "Workspace", "Requested")

	for _, approval := range approvals {
		if approval.Status == statusPending {
			userName := fmt.Sprintf("%s %s", approval.WorkspaceMembership.User.Name, approval.WorkspaceMembership.User.Surname)
			orgSlug := approval.WorkspaceMembership.Workspace.Organization.Slug
			workspaceSlug := approval.WorkspaceMembership.Workspace.Slug
			workspaceName := fmt.Sprintf("%s/%s", orgSlug, workspaceSlug)

			t.AddRow(
				fmt.Sprintf("%d", approval.ID),
				truncateString(userName, maxDisplayLength),
				truncateString(approval.Device.Name, maxDisplayLength),
				truncateString(workspaceName, maxDisplayLength),
				truncateString(formatTime(approval.InsertedAt), maxDisplayLength),
			)
		}
	}

	err = t.Render()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("💡 Use 'initiat device approve --all' to approve all pending devices")
	fmt.Println("💡 Use 'initiat device approve --id <id>' to approve a specific device")

	return nil
}

func runApproveDevice(cmd *cobra.Command, args []string) error {
	apiClient := client.New()

	if approveAll {
		return runApproveAllDevices(apiClient)
	}

	if approvalID == "" {
		return fmt.Errorf("❌ Please specify either --all or --id <approval-id>")
	}

	return runApproveSingleDevice(apiClient, approvalID)
}

func runRejectDevice(cmd *cobra.Command, args []string) error {
	apiClient := client.New()

	if rejectAll {
		return runRejectAllDevices(apiClient)
	}

	if approvalID == "" {
		return fmt.Errorf("❌ Please specify either --all or --id <approval-id>")
	}

	return runRejectSingleDevice(apiClient, approvalID)
}

func runShowApproval(cmd *cobra.Command, args []string) error {
	apiClient := client.New()

	approval, err := apiClient.GetDeviceApproval(approvalID)
	if err != nil {
		return fmt.Errorf("❌ Failed to get device approval: %w", err)
	}

	fmt.Println("📋 Device Approval Details")
	fmt.Println()
	fmt.Printf("User: %s %s (%s)\n",
		approval.WorkspaceMembership.User.Name,
		approval.WorkspaceMembership.User.Surname,
		approval.WorkspaceMembership.User.Email)
	fmt.Printf("Device: %s (ID: %d)\n", approval.Device.Name, approval.Device.ID)
	fmt.Printf("Workspace: %s / %s (%s/%s)\n",
		approval.WorkspaceMembership.Workspace.Organization.Name,
		approval.WorkspaceMembership.Workspace.Name,
		approval.WorkspaceMembership.Workspace.Organization.Slug,
		approval.WorkspaceMembership.Workspace.Slug)
	fmt.Printf("Requested: %s\n", formatTime(approval.InsertedAt))
	fmt.Printf("Status: %s\n", approval.Status)

	if approval.ApprovedByUser != nil {
		fmt.Printf("Approved by: %s %s (%s)\n",
			approval.ApprovedByUser.Name,
			approval.ApprovedByUser.Surname,
			approval.ApprovedByUser.Email)
	}

	fmt.Println()
	fmt.Println("🔑 Device Public Keys:")
	fmt.Printf("  Ed25519: %s... (for signing)\n", truncateString(approval.Device.PublicKeyEd25519, maxKeyDisplayLength))
	fmt.Printf("  X25519: %s... (for encryption)\n", truncateString(approval.Device.PublicKeyX25519, maxKeyDisplayLength))

	return nil
}

func runApproveAllDevices(apiClient *client.Client) error {
	approvals, err := apiClient.ListDeviceApprovals()
	if err != nil {
		return fmt.Errorf("❌ Failed to list device approvals: %w", err)
	}

	pendingApprovals := filterPendingApprovals(approvals)
	if len(pendingApprovals) == 0 {
		fmt.Println("📋 No pending device approvals found")
		return nil
	}

	fmt.Printf("🔐 Approving all pending devices...\n\n")
	fmt.Printf("Found %d pending approvals:\n", len(pendingApprovals))

	workspaceKeys := collectWorkspaceKeys(pendingApprovals)
	if len(workspaceKeys) == 0 {
		return fmt.Errorf("❌ No workspace keys found")
	}

	fmt.Println()
	successCount := approveDevicesBatch(apiClient, pendingApprovals, workspaceKeys)

	fmt.Printf("✅ Approved %d devices successfully!\n", successCount)
	fmt.Println("   All approved devices can now access their respective workspace secrets")

	return nil
}

func filterPendingApprovals(approvals []types.DeviceApproval) []types.DeviceApproval {
	pendingApprovals := make([]types.DeviceApproval, 0)
	for _, approval := range approvals {
		if approval.Status == statusPending {
			pendingApprovals = append(pendingApprovals, approval)
		}
	}
	return pendingApprovals
}

func collectWorkspaceKeys(pendingApprovals []types.DeviceApproval) map[string][]byte {
	workspaceKeys := make(map[string][]byte)

	for _, approval := range pendingApprovals {
		workspaceSlug := buildWorkspaceSlug(approval)

		fmt.Printf("  • %s (%s) - %s %s\n",
			approval.Device.Name,
			workspaceSlug,
			approval.WorkspaceMembership.User.Name,
			approval.WorkspaceMembership.User.Surname)

		if _, exists := workspaceKeys[workspaceSlug]; !exists {
			key, err := getWorkspaceKeyForApproval(workspaceSlug)
			if err != nil {
				fmt.Printf("❌ Failed to get workspace key for %s: %v\n", workspaceSlug, err)
				continue
			}
			workspaceKeys[workspaceSlug] = key
		}
	}

	return workspaceKeys
}

func approveDevicesBatch(
	apiClient *client.Client,
	pendingApprovals []types.DeviceApproval,
	workspaceKeys map[string][]byte,
) int {
	successCount := 0

	for _, approval := range pendingApprovals {
		workspaceSlug := buildWorkspaceSlug(approval)
		workspaceKey := workspaceKeys[workspaceSlug]

		devicePublicKey, err := encoding.Decode(approval.Device.PublicKeyX25519)
		if err != nil {
			fmt.Printf("❌ Failed to decode device public key for %s: %v\n", approval.Device.Name, err)
			continue
		}

		wrappedKey, err := encoding.WrapWorkspaceKey(workspaceKey, devicePublicKey)
		if err != nil {
			fmt.Printf("❌ Failed to wrap workspace key for %s: %v\n", approval.Device.Name, err)
			continue
		}

		_, err = apiClient.ApproveDevice(fmt.Sprintf("%d", approval.ID), wrappedKey)
		if err != nil {
			fmt.Printf("❌ Failed to approve %s: %v\n", approval.Device.Name, err)
			continue
		}

		successCount++
	}

	return successCount
}

func runApproveSingleDevice(apiClient *client.Client, approvalID string) error {
	approval, err := apiClient.GetDeviceApproval(approvalID)
	if err != nil {
		return fmt.Errorf("❌ Failed to get device approval: %w", err)
	}

	if approval.Status != "pending" {
		return fmt.Errorf("❌ Device approval is not pending (status: %s)", approval.Status)
	}

	workspaceSlug := buildWorkspaceSlug(*approval)

	fmt.Printf("🔐 Approving device for %s...\n", approval.Device.Name)

	workspaceKey, err := getWorkspaceKeyForApproval(workspaceSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	devicePublicKey, err := encoding.Decode(approval.Device.PublicKeyX25519)
	if err != nil {
		return fmt.Errorf("❌ Failed to decode device public key: %w", err)
	}

	wrappedKey, err := encoding.WrapWorkspaceKey(workspaceKey, devicePublicKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to wrap workspace key: %w", err)
	}

	_, err = apiClient.ApproveDevice(approvalID, wrappedKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to approve device: %w", err)
	}

	fmt.Println("✅ Device approved successfully!")
	fmt.Println("   Device can now access workspace secrets")

	return nil
}

func runRejectAllDevices(apiClient *client.Client) error {
	approvals, err := apiClient.ListDeviceApprovals()
	if err != nil {
		return fmt.Errorf("❌ Failed to list device approvals: %w", err)
	}

	pendingApprovals := make([]types.DeviceApproval, 0)
	for _, approval := range approvals {
		if approval.Status == statusPending {
			pendingApprovals = append(pendingApprovals, approval)
		}
	}

	if len(pendingApprovals) == 0 {
		fmt.Println("📋 No pending device approvals found")
		return nil
	}

	fmt.Printf("❌ Rejecting all pending devices...\n\n")
	fmt.Printf("Found %d pending approvals to reject\n", len(pendingApprovals))

	successCount := 0

	for _, approval := range pendingApprovals {
		_, err := apiClient.RejectDevice(fmt.Sprintf("%d", approval.ID))
		if err != nil {
			fmt.Printf("❌ Failed to reject %s: %v\n", approval.Device.Name, err)
			continue
		}

		successCount++
	}

	fmt.Printf("❌ Rejected %d devices\n", successCount)
	fmt.Println("   Users will need to request approval again")

	return nil
}

func runRejectSingleDevice(apiClient *client.Client, approvalID string) error {
	approval, err := apiClient.GetDeviceApproval(approvalID)
	if err != nil {
		return fmt.Errorf("❌ Failed to get device approval: %w", err)
	}

	if approval.Status != "pending" {
		return fmt.Errorf("❌ Device approval is not pending (status: %s)", approval.Status)
	}

	_, err = apiClient.RejectDevice(approvalID)
	if err != nil {
		return fmt.Errorf("❌ Failed to reject device: %w", err)
	}

	fmt.Printf("❌ Device rejected for %s\n", approval.Device.Name)
	fmt.Println("   User will need to request approval again")

	return nil
}

func buildWorkspaceSlug(approval types.DeviceApproval) string {
	workspaceSlug := fmt.Sprintf("%s/%s",
		approval.WorkspaceMembership.Workspace.Organization.Slug,
		approval.WorkspaceMembership.Workspace.Slug)

	if approval.WorkspaceMembership.Workspace.Organization.Slug == "" {
		workspaceSlug = approval.WorkspaceMembership.Workspace.Slug
	}

	return workspaceSlug
}

func parseWorkspaceSlug(compositeSlug string) (string, string, error) {
	parts := strings.Split(compositeSlug, "/")
	const expectedParts = 2
	if len(parts) != expectedParts {
		return "", "", fmt.Errorf(
			"invalid workspace slug format: expected 'org-slug/workspace-slug', got '%s'",
			compositeSlug,
		)
	}
	return parts[0], parts[1], nil
}

func getWorkspaceKeyForApproval(compositeSlug string) ([]byte, error) {
	store := storage.New()

	orgSlug, workspaceSlug, err := parseWorkspaceSlug(compositeSlug)
	if err != nil {
		return nil, err
	}

	apiClient := client.New()
	wrappedKey, err := apiClient.GetWrappedWorkspaceKey(orgSlug, workspaceSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wrapped workspace key: %w", err)
	}

	devicePrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get device private key: %w", err)
	}

	workspaceKey, err := encoding.UnwrapWorkspaceKey(wrappedKey, devicePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap workspace key: %w", err)
	}

	if len(workspaceKey) != encoding.WorkspaceKeySize {
		return nil, fmt.Errorf("invalid workspace key size: %d bytes (expected %d)",
			len(workspaceKey), encoding.WorkspaceKeySize)
	}

	return workspaceKey, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= minTruncateLength {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("Jan 2 15:04")
}
