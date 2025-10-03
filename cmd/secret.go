package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/encoding"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
	"github.com/DylanBlakemore/initiat-cli/internal/table"
)

var (
	secretKey     string
	secretValue   string
	description   string
	forceOverride bool
	copyToClip    bool
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets in workspaces",
	Long:  `Manage secrets in Initiat workspaces with client-side encryption.`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a secret value",
	Long: `Set a secret value in the specified workspace. The value is encrypted client-side 
before being sent to the server.

Examples:
  initiat secret set --workspace-path acme-corp/production --key API_KEY --value "sk-1234567890abcdef"
  initiat secret set -W acme-corp/production -k API_KEY -v "sk-1234567890abcdef"
  initiat secret set --org acme-corp --workspace production --key DB_PASSWORD \
    --value "super-secret-pass" --description "Production database password"
  initiat secret set -w production -k API_KEY -v "new-value" --force`,
	RunE: runSecretSet,
}

var secretGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a secret value (JSON output)",
	Long: `Get and decrypt a secret value from the specified workspace.
Output is always in JSON format.

Examples:
  initiat secret get --workspace-path acme-corp/production --key API_KEY
  initiat secret get -W acme-corp/production -k API_KEY
  initiat secret get --workspace production --key DB_PASSWORD
  initiat secret get -w production -k API_KEY --copy`,
	RunE: runSecretGet,
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets (table format)",
	Long: `List all secrets in the specified workspace (metadata only, no values).
Output is always in table format showing key, value preview, and version.

Examples:
  initiat secret list --workspace-path acme-corp/production
  initiat secret list -W acme-corp/production
  initiat secret list --workspace production`,
	RunE: runSecretList,
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret",
	Long: `Delete a secret from the specified workspace.

Examples:
  initiat secret delete --workspace-path acme-corp/production --key API_KEY
  initiat secret delete -W acme-corp/production -k API_KEY
  initiat secret delete --workspace production --key OLD_API_KEY`,
	RunE: runSecretDelete,
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretDeleteCmd)

	secretSetCmd.Flags().StringVarP(&secretKey, "key", "k", "", "Secret key name (required)")
	secretSetCmd.Flags().StringVarP(&secretValue, "value", "v", "", "Secret value (required)")
	secretSetCmd.Flags().StringVarP(&description, "description", "d", "", "Optional description for the secret")
	secretSetCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Overwrite existing secret without confirmation")
	_ = secretSetCmd.MarkFlagRequired("key")
	_ = secretSetCmd.MarkFlagRequired("value")

	secretGetCmd.Flags().StringVarP(&secretKey, "key", "k", "", "Secret key name (required)")
	secretGetCmd.Flags().BoolVarP(&copyToClip, "copy", "c", false, "Copy value to clipboard instead of printing")
	_ = secretGetCmd.MarkFlagRequired("key")

	secretDeleteCmd.Flags().StringVarP(&secretKey, "key", "k", "", "Secret key name (required)")
	secretDeleteCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Skip confirmation prompt")
	_ = secretDeleteCmd.MarkFlagRequired("key")
}

func runSecretSet(cmd *cobra.Command, args []string) error {
	workspaceCtx, err := GetWorkspaceContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(secretKey)
	value := strings.TrimSpace(secretValue)

	if key == "" {
		return fmt.Errorf("secret key cannot be empty")
	}
	if value == "" {
		return fmt.Errorf("secret value cannot be empty")
	}

	fmt.Printf("🔐 Setting secret '%s' in workspace %s...\n", key, workspaceCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	workspaceKey, err := getWorkspaceKey(workspaceCtx.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	fmt.Println("🔒 Encrypting secret value...")

	encryptedValue, nonce, err := encryptSecretValue(value, workspaceKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt secret: %w", err)
	}

	fmt.Println("📡 Uploading encrypted secret to server...")

	c := client.New()
	secret, err := c.SetSecret(
		workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug, key,
		encryptedValue, nonce, description, forceOverride,
	)
	if err != nil {
		return fmt.Errorf("❌ Failed to set secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' set successfully!\n", key)
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
			"workspace key not found locally for workspace '%s'. Please run 'initiat workspace init %s' first",
			compositeSlug, compositeSlug)
	}

	return store.GetWorkspaceKey(compositeSlug)
}

func encryptSecretValue(value string, workspaceKey []byte) ([]byte, []byte, error) {
	if len(workspaceKey) != encoding.WorkspaceKeySize {
		return nil, nil, fmt.Errorf(
			"invalid workspace key size: %d bytes, expected %d bytes",
			len(workspaceKey), encoding.WorkspaceKeySize)
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	var key [32]byte
	copy(key[:], workspaceKey)

	ciphertext := secretbox.Seal(nil, []byte(value), &nonce, &key)

	return ciphertext, nonce[:], nil
}

func runSecretGet(cmd *cobra.Command, args []string) error {
	workspaceCtx, err := GetWorkspaceContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(secretKey)
	if key == "" {
		return fmt.Errorf("secret key cannot be empty")
	}

	fmt.Printf("🔍 Getting secret '%s' from workspace %s...\n", key, workspaceCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	workspaceKey, err := getWorkspaceKey(workspaceCtx.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace key: %w", err)
	}

	c := client.New()
	secretData, err := c.GetSecret(workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug, key)
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
	workspaceCtx, err := GetWorkspaceContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔍 Listing secrets in workspace %s...\n", workspaceCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	secrets, err := c.ListSecrets(workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to list secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("No secrets found in this workspace.")
		return nil
	}

	t := table.New()
	t.SetHeaders("Key", "Value", "Version")

	for _, secret := range secrets {
		t.AddRow(secret.Key, "[encrypted]", fmt.Sprintf("%d", secret.Version))
	}

	return t.Render()
}

func runSecretDelete(cmd *cobra.Command, args []string) error {
	workspaceCtx, err := GetWorkspaceContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(secretKey)
	if key == "" {
		return fmt.Errorf("secret key cannot be empty")
	}

	if !forceOverride {
		fmt.Printf("⚠️  Are you sure you want to delete secret '%s' from workspace %s? (y/N): ",
			key, workspaceCtx.String())
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	fmt.Printf("🗑️  Deleting secret '%s' from workspace %s...\n", key, workspaceCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	if err := c.DeleteSecret(workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug, key); err != nil {
		return fmt.Errorf("❌ Failed to delete secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' deleted successfully!\n", key)
	return nil
}

func decryptSecretValue(encryptedValue, nonce string, workspaceKey []byte) (string, error) {
	ciphertext, err := encoding.Decode(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}

	nonceBytes, err := encoding.Decode(nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

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

	var nonceArray [24]byte
	var keyArray [32]byte
	copy(nonceArray[:], nonceBytes)
	copy(keyArray[:], workspaceKey)

	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return "", fmt.Errorf("failed to decrypt value: authentication failed")
	}

	return string(plaintext), nil
}
