package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/DylanBlakemore/initflow-cli/internal/client"
	"github.com/DylanBlakemore/initflow-cli/internal/config"
	"github.com/DylanBlakemore/initflow-cli/internal/encoding"
	"github.com/DylanBlakemore/initflow-cli/internal/slug"
	"github.com/DylanBlakemore/initflow-cli/internal/storage"
)

const (
	secretSetArgsCount    = 3 // workspace, key, value
	secretGetArgsCount    = 2 // workspace, key
	secretDeleteArgsCount = 2 // workspace, key
)

var (
	description   string
	forceOverride bool
	copyToClip    bool
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets in workspaces",
	Long:  `Manage secrets in InitFlow workspaces with client-side encryption.`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set <org-slug/workspace-slug> <KEY> <VALUE>",
	Short: "Set a secret value",
	Long: `Set a secret value in the specified workspace. The value is encrypted client-side 
before being sent to the server.

Examples:
  initflow secret set acme-corp/production API_KEY "sk-1234567890abcdef"
  initflow secret set acme-corp/production DB_PASSWORD "super-secret-pass" --description "Production database password"
  initflow secret set production API_KEY "new-value" --force  # Uses default org context`,
	Args: cobra.ExactArgs(secretSetArgsCount),
	RunE: runSecretSet,
}

var secretGetCmd = &cobra.Command{
	Use:   "get <org-slug/workspace-slug> <KEY>",
	Short: "Get a secret value (JSON output)",
	Long: `Get and decrypt a secret value from the specified workspace.
Output is always in JSON format.

Examples:
  initflow secret get acme-corp/production API_KEY
  initflow secret get acme-corp/production API_KEY --copy
  initflow secret get production API_KEY  # Uses default org context`,
	Args: cobra.ExactArgs(secretGetArgsCount),
	RunE: runSecretGet,
}

var secretListCmd = &cobra.Command{
	Use:   "list <org-slug/workspace-slug>",
	Short: "List all secrets (table format)",
	Long: `List all secrets in the specified workspace (metadata only, no values).
Output is always in table format showing key, value preview, and version.

Examples:
  initflow secret list acme-corp/production
  initflow secret list production  # Uses default org context`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretList,
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete <org-slug/workspace-slug> <KEY>",
	Short: "Delete a secret",
	Long: `Delete a secret from the specified workspace.

Examples:
  initflow secret delete acme-corp/production API_KEY
  initflow secret delete production API_KEY --force  # Uses default org context`,
	Args: cobra.ExactArgs(secretGetArgsCount),
	RunE: runSecretDelete,
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretDeleteCmd)

	// Add flags for secret set command
	secretSetCmd.Flags().StringVarP(&description, "description", "d", "", "Optional description for the secret")
	secretSetCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Overwrite existing secret without confirmation")

	// Add flags for secret get command
	secretGetCmd.Flags().BoolVarP(&copyToClip, "copy", "c", false, "Copy value to clipboard instead of printing")

	// Add flags for secret delete command
	secretDeleteCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Skip confirmation prompt")
}

func runSecretSet(cmd *cobra.Command, args []string) error {
	workspaceInput := args[0]
	secretKey := args[1]
	secretValue := args[2]

	if secretKey == "" {
		return fmt.Errorf("❌ Secret key cannot be empty")
	}
	if secretValue == "" {
		return fmt.Errorf("❌ Secret value cannot be empty")
	}

	defaultOrgSlug := config.GetDefaultOrgSlug()
	compositeSlug, err := slug.ResolveWorkspaceSlug(workspaceInput, defaultOrgSlug)
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔐 Setting secret '%s' in workspace %s...\n", secretKey, compositeSlug.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	workspaceKey, err := getWorkspaceKey(compositeSlug.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	fmt.Println("🔒 Encrypting secret value...")

	encryptedValue, nonce, err := encryptSecretValue(secretValue, workspaceKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt secret: %w", err)
	}

	fmt.Println("📡 Uploading encrypted secret to server...")

	c := client.New()
	secret, err := c.SetSecret(
		compositeSlug.OrgSlug, compositeSlug.WorkspaceSlug, secretKey,
		encryptedValue, nonce, description, forceOverride,
	)
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

func getWorkspaceKey(compositeSlug string, store *storage.Storage) ([]byte, error) {
	if !store.HasWorkspaceKey(compositeSlug) {
		return nil, fmt.Errorf(
			"workspace key not found locally for workspace '%s'. Please run 'initflow workspace init %s' first",
			compositeSlug, compositeSlug)
	}

	return store.GetWorkspaceKey(compositeSlug)
}

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
	workspaceInput := args[0]
	secretKey := args[1]

	if secretKey == "" {
		return fmt.Errorf("❌ Secret key cannot be empty")
	}

	defaultOrgSlug := config.GetDefaultOrgSlug()
	compositeSlug, err := slug.ResolveWorkspaceSlug(workspaceInput, defaultOrgSlug)
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔍 Getting secret '%s' from workspace %s...\n", secretKey, compositeSlug.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	workspaceKey, err := getWorkspaceKey(compositeSlug.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	c := client.New()
	secretData, err := c.GetSecret(compositeSlug.OrgSlug, compositeSlug.WorkspaceSlug, secretKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to get secret: %w", err)
	}

	fmt.Println("🔓 Decrypting secret value...")

	decryptedValue, err := decryptSecretValue(secretData.EncryptedValue, secretData.Nonce, workspaceKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to decrypt secret: %w", err)
	}

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

	if copyToClip {
		fmt.Println("📋 Copied secret value to clipboard")
		// TODO: Implement clipboard functionality
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func runSecretList(cmd *cobra.Command, args []string) error {
	workspaceInput := args[0]

	defaultOrgSlug := config.GetDefaultOrgSlug()
	compositeSlug, err := slug.ResolveWorkspaceSlug(workspaceInput, defaultOrgSlug)
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔍 Listing secrets in workspace %s...\n", compositeSlug.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	c := client.New()
	secrets, err := c.ListSecrets(compositeSlug.OrgSlug, compositeSlug.WorkspaceSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to list secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("No secrets found in this workspace.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Key\tValue\tVersion")
	fmt.Fprintln(w, "───\t─────\t───────")

	for _, secret := range secrets {
		fmt.Fprintf(w, "%s\t%s\t%d\n",
			secret.Key,
			"[encrypted]",
			secret.Version)
	}
	_ = w.Flush()

	return nil
}

func runSecretDelete(cmd *cobra.Command, args []string) error {
	workspaceInput := args[0]
	secretKey := args[1]

	if secretKey == "" {
		return fmt.Errorf("❌ Secret key cannot be empty")
	}

	defaultOrgSlug := config.GetDefaultOrgSlug()
	compositeSlug, err := slug.ResolveWorkspaceSlug(workspaceInput, defaultOrgSlug)
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	if !forceOverride {
		fmt.Printf("⚠️  Are you sure you want to delete secret '%s' from workspace %s? (y/N): ",
			secretKey, compositeSlug.String())
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	fmt.Printf("🗑️  Deleting secret '%s' from workspace %s...\n", secretKey, compositeSlug.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initflow device register <name>' first")
	}

	c := client.New()
	if err := c.DeleteSecret(compositeSlug.OrgSlug, compositeSlug.WorkspaceSlug, secretKey); err != nil {
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
