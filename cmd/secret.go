package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/DylanBlakemore/initflow-cli/internal/client"
	"github.com/DylanBlakemore/initflow-cli/internal/encoding"
	"github.com/DylanBlakemore/initflow-cli/internal/storage"
)

var (
	workspaceID   int
	description   string
	forceOverride bool
	outputFormat  string
	copyToClip    bool
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets in workspaces",
	Long:  `Manage secrets in InitFlow workspaces with client-side encryption.`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set <KEY> <VALUE>",
	Short: "Set a secret value",
	Long: `Set a secret value in the specified workspace. The value is encrypted client-side 
before being sent to the server.

Examples:
  initflow secret set API_KEY "sk-1234567890abcdef" --workspace 42
  initflow secret set DB_PASSWORD "super-secret-pass" --workspace 42 --description "Production database password"
  initflow secret set API_KEY "new-value" --workspace 42 --force`,
	Args: cobra.ExactArgs(2), //nolint:mnd // Command requires exactly 2 arguments: key and value
	RunE: runSecretSet,
}

var secretGetCmd = &cobra.Command{
	Use:   "get <KEY>",
	Short: "Get a secret value",
	Long: `Get and decrypt a secret value from the specified workspace.

Examples:
  initflow secret get API_KEY --workspace 42
  initflow secret get API_KEY --workspace 42 --copy
  initflow secret get API_KEY --workspace 42 --output json
  initflow secret get API_KEY --workspace 42 --output env`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretGet,
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Long: `List all secrets in the specified workspace (metadata only, no values).

Examples:
  initflow secret list --workspace 42
  initflow secret list --workspace 42 --format json
  initflow secret list --workspace 42 --format simple`,
	RunE: runSecretList,
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete <KEY>",
	Short: "Delete a secret",
	Long: `Delete a secret from the specified workspace.

Examples:
  initflow secret delete API_KEY --workspace 42
  initflow secret delete API_KEY --workspace 42 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretDelete,
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretDeleteCmd)

	// Add flags for secret set command
	secretSetCmd.Flags().IntVarP(&workspaceID, "workspace", "w", 0, "Workspace ID (required)")
	secretSetCmd.Flags().StringVarP(&description, "description", "d", "", "Optional description for the secret")
	secretSetCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Overwrite existing secret without confirmation")
	_ = secretSetCmd.MarkFlagRequired("workspace")

	// Add flags for secret get command
	secretGetCmd.Flags().IntVarP(&workspaceID, "workspace", "w", 0, "Workspace ID (required)")
	secretGetCmd.Flags().StringVarP(&outputFormat, "output", "o", "value", "Output format (value|json|env)")
	secretGetCmd.Flags().BoolVarP(&copyToClip, "copy", "c", false, "Copy value to clipboard instead of printing")
	_ = secretGetCmd.MarkFlagRequired("workspace")

	// Add flags for secret list command
	secretListCmd.Flags().IntVarP(&workspaceID, "workspace", "w", 0, "Workspace ID (required)")
	secretListCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format (table|json|simple)")
	_ = secretListCmd.MarkFlagRequired("workspace")

	// Add flags for secret delete command
	secretDeleteCmd.Flags().IntVarP(&workspaceID, "workspace", "w", 0, "Workspace ID (required)")
	secretDeleteCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Skip confirmation prompt")
	_ = secretDeleteCmd.MarkFlagRequired("workspace")
}

func runSecretSet(cmd *cobra.Command, args []string) error {
	secretKey := args[0]
	secretValue := args[1]

	// Validate inputs
	if secretKey == "" {
		return fmt.Errorf("❌ Secret key cannot be empty")
	}
	if secretValue == "" {
		return fmt.Errorf("❌ Secret value cannot be empty")
	}
	if workspaceID <= 0 {
		return fmt.Errorf("❌ Invalid workspace ID: %d", workspaceID)
	}

	fmt.Printf("🔐 Setting secret '%s' in workspace %d...\n", secretKey, workspaceID)

	// Check device registration
	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	// Get workspace key
	workspaceKey, err := getWorkspaceKeyByID(workspaceID, store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	fmt.Println("🔒 Encrypting secret value...")

	// Encrypt the secret value
	encryptedValue, nonce, err := encryptSecretValue(secretValue, workspaceKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt secret: %w", err)
	}

	fmt.Println("📡 Uploading encrypted secret to server...")

	// Create API client and submit secret
	c := client.New()
	secret, err := c.SetSecret(workspaceID, secretKey, encryptedValue, nonce, description, forceOverride)
	if err != nil {
		return fmt.Errorf("❌ Failed to set secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' set successfully!\n", secretKey)
	fmt.Printf("   Version: %d\n", secret.Version)
	fmt.Printf("   Updated: %s\n", secret.UpdatedAt)
	if secret.CreatedByDevice.Name != "" {
		fmt.Printf("   Created by: %s\n", secret.CreatedByDevice.Name)
	}

	return nil
}

// getWorkspaceKeyByID retrieves the workspace key for a given workspace ID
// It first tries to get it from local storage, and if not found, fetches workspace info to get the slug
func getWorkspaceKeyByID(workspaceID int, store *storage.Storage) ([]byte, error) {
	// First, we need to get the workspace slug from the workspace ID
	// We'll fetch all workspaces and find the matching one
	c := client.New()
	workspaces, err := c.ListWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	var workspaceSlug string
	var found bool
	for _, workspace := range workspaces {
		if workspace.ID == workspaceID {
			workspaceSlug = workspace.Slug
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("workspace with ID %d not found or not accessible", workspaceID)
	}

	// Check if we have the workspace key locally
	if !store.HasWorkspaceKey(workspaceSlug) {
		return nil, fmt.Errorf(
			"workspace key not found locally for workspace '%s'. Please run 'initflow workspace init %s' first",
			workspaceSlug, workspaceSlug)
	}

	return store.GetWorkspaceKey(workspaceSlug)
}

// encryptSecretValue encrypts a secret value using the workspace key with NaCl secretbox
func encryptSecretValue(value string, workspaceKey []byte) ([]byte, []byte, error) {
	// Validate workspace key size
	if len(workspaceKey) != encoding.WorkspaceKeySize {
		return nil, nil, fmt.Errorf(
			"invalid workspace key size: %d bytes, expected %d bytes",
			len(workspaceKey), encoding.WorkspaceKeySize)
	}

	// Generate random nonce (24 bytes for XSalsa20Poly1305)
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Convert workspace key to fixed-size array
	var key [32]byte
	copy(key[:], workspaceKey)

	// Encrypt the value using NaCl secretbox (XSalsa20Poly1305)
	ciphertext := secretbox.Seal(nil, []byte(value), &nonce, &key)

	return ciphertext, nonce[:], nil
}

func runSecretGet(cmd *cobra.Command, args []string) error {
	secretKey := args[0]

	if secretKey == "" {
		return fmt.Errorf("❌ Secret key cannot be empty")
	}
	if workspaceID <= 0 {
		return fmt.Errorf("❌ Invalid workspace ID: %d", workspaceID)
	}

	fmt.Printf("🔍 Getting secret '%s' from workspace %d...\n", secretKey, workspaceID)

	// Check device registration
	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	// Get workspace key
	workspaceKey, err := getWorkspaceKeyByID(workspaceID, store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	// Get encrypted secret from server
	c := client.New()
	secretData, err := c.GetSecret(workspaceID, secretKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to get secret: %w", err)
	}

	fmt.Println("🔓 Decrypting secret value...")

	// Decrypt the secret value
	decryptedValue, err := decryptSecretValue(secretData.EncryptedValue, secretData.Nonce, workspaceKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to decrypt secret: %w", err)
	}

	// Output in requested format
	switch outputFormat {
	case "value":
		fmt.Println(decryptedValue)
	case "json":
		output := map[string]interface{}{
			"key":               secretData.Key,
			"value":             decryptedValue,
			"version":           secretData.Version,
			"workspace_id":      secretData.WorkspaceID,
			"updated_at":        secretData.UpdatedAt,
			"created_by_device": secretData.CreatedByDevice.Name,
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("❌ Failed to format JSON output: %w", err)
		}
		fmt.Println(string(jsonData))
	case "env":
		fmt.Printf("export %s=\"%s\"\n", secretKey, decryptedValue)
	default:
		return fmt.Errorf("❌ Invalid output format: %s (valid: value, json, env)", outputFormat)
	}

	return nil
}

func runSecretList(cmd *cobra.Command, args []string) error {
	if workspaceID <= 0 {
		return fmt.Errorf("❌ Invalid workspace ID: %d", workspaceID)
	}

	fmt.Printf("🔍 Listing secrets in workspace %d...\n", workspaceID)

	// Check device registration
	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	// List secrets from server
	c := client.New()
	secrets, err := c.ListSecrets(workspaceID)
	if err != nil {
		return fmt.Errorf("❌ Failed to list secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("No secrets found in this workspace.")
		return nil
	}

	// Output in requested format
	switch outputFormat {
	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "Key\tVersion\tUpdated\tCreated By")
		fmt.Fprintln(w, "───\t───────\t───────\t──────────")

		for _, secret := range secrets {
			updatedTime, _ := time.Parse(time.RFC3339, secret.UpdatedAt)
			timeAgo := formatTimeAgo(updatedTime)

			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
				secret.Key,
				secret.Version,
				timeAgo,
				secret.CreatedByDevice.Name)
		}
		_ = w.Flush()

	case "simple":
		for _, secret := range secrets {
			fmt.Println(secret.Key)
		}

	case "json":
		output := make([]map[string]interface{}, len(secrets))
		for i, secret := range secrets {
			output[i] = map[string]interface{}{
				"key":               secret.Key,
				"version":           secret.Version,
				"updated_at":        secret.UpdatedAt,
				"created_by_device": secret.CreatedByDevice.Name,
			}
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("❌ Failed to format JSON output: %w", err)
		}
		fmt.Println(string(jsonData))

	default:
		return fmt.Errorf("❌ Invalid format: %s (valid: table, simple, json)", outputFormat)
	}

	return nil
}

func runSecretDelete(cmd *cobra.Command, args []string) error {
	secretKey := args[0]

	if secretKey == "" {
		return fmt.Errorf("❌ Secret key cannot be empty")
	}
	if workspaceID <= 0 {
		return fmt.Errorf("❌ Invalid workspace ID: %d", workspaceID)
	}

	// Confirmation prompt unless --force is used
	if !forceOverride {
		fmt.Printf("⚠️  Are you sure you want to delete secret '%s' from workspace %d? (y/N): ", secretKey, workspaceID)
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	fmt.Printf("🗑️  Deleting secret '%s' from workspace %d...\n", secretKey, workspaceID)

	// Check device registration
	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	// Delete secret from server
	c := client.New()
	if err := c.DeleteSecret(workspaceID, secretKey); err != nil {
		return fmt.Errorf("❌ Failed to delete secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' deleted successfully!\n", secretKey)
	return nil
}

// decryptSecretValue decrypts a secret value using NaCl secretbox (XSalsa20Poly1305)
func decryptSecretValue(encryptedValue, nonce string, workspaceKey []byte) (string, error) {
	// Decode the encrypted value and nonce
	ciphertext, err := encoding.Decode(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}

	nonceBytes, err := encoding.Decode(nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Validate sizes
	if len(nonceBytes) != encoding.SecretboxNonceSize {
		return "", fmt.Errorf(
			"invalid nonce size: got %d bytes, expected %d bytes",
			len(nonceBytes), encoding.SecretboxNonceSize)
	}

	if len(workspaceKey) != encoding.WorkspaceKeySize {
		return "", fmt.Errorf(
			"invalid workspace key size: got %d bytes, expected %d bytes",
			len(workspaceKey), encoding.WorkspaceKeySize)
	}

	// Convert to fixed-size arrays for NaCl secretbox
	var nonceArray [24]byte
	var keyArray [32]byte
	copy(nonceArray[:], nonceBytes)
	copy(keyArray[:], workspaceKey)

	// Decrypt the value using NaCl secretbox (XSalsa20Poly1305)
	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return "", fmt.Errorf("failed to decrypt value: authentication failed")
	}

	return string(plaintext), nil
}

// formatTimeAgo formats a time as a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	const (
		hoursPerDay  = 24
		daysPerMonth = 30
	)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < hoursPerDay*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < daysPerMonth*hoursPerDay*time.Hour:
		days := int(diff.Hours() / hoursPerDay)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02")
	}
}
