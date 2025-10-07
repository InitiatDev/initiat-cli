package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.design/x/clipboard"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/export"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/table"
	"github.com/InitiatDev/initiat-cli/internal/validation"
)

var (
	secretValue   string
	description   string
	forceOverride bool
	copyToClip    bool
	outputFile    string
	copyKeyValue  bool
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets in projects",
	Long:  `Manage secrets in Initiat projects with client-side encryption.`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set <secret-key>",
	Short: "Set a secret value",
	Long: `Set a secret value in the specified project. The value is encrypted client-side 
before being sent to the server.

Examples:
  initiat secret set API_KEY --project-path acme-corp/production --value "sk-1234567890abcdef"
  initiat secret set API_KEY -P acme-corp/production -v "sk-1234567890abcdef"
  initiat secret set DB_PASSWORD --org acme-corp --project production \
    --value "super-secret-pass" --description "Production database password"
  initiat secret set API_KEY -p production -v "new-value" --force`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretSet,
}

var secretGetCmd = &cobra.Command{
	Use:   "get <secret-key>",
	Short: "Get a secret value (JSON output)",
	Long: `Get and decrypt a secret value from the specified project.
Output is always in JSON format.

Examples:
  initiat secret get API_KEY --project-path acme-corp/production
  initiat secret get API_KEY -P acme-corp/production
  initiat secret get DB_PASSWORD --project production
  initiat secret get API_KEY -p production --copy`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretGet,
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets (table format)",
	Long: `List all secrets in the specified project (metadata only, no values).
Output is always in table format showing key, value preview, and version.

Examples:
  initiat secret list --project-path acme-corp/production
  initiat secret list -P acme-corp/production
  initiat secret list --project production`,
	RunE: runSecretList,
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete <secret-key>",
	Short: "Delete a secret",
	Long: `Delete a secret from the specified project.

Examples:
  initiat secret delete API_KEY --project-path acme-corp/production
  initiat secret delete API_KEY -P acme-corp/production
  initiat secret delete OLD_API_KEY --project production`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretDelete,
}

var secretExportCmd = &cobra.Command{
	Use:   "export <secret-key>",
	Short: "Export a secret to a file",
	Long: `Export a secret value to a file. Creates directories if needed and handles overwrite prompts.

Examples:
  initiat secret export API_KEY -o .env
  initiat secret export API_KEY -o config/secrets.env`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretExport,
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretDeleteCmd)
	secretCmd.AddCommand(secretExportCmd)

	secretSetCmd.Flags().StringVarP(&secretValue, "value", "v", "", "Secret value (required)")
	secretSetCmd.Flags().StringVarP(&description, "description", "d", "", "Optional description for the secret")
	secretSetCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Overwrite existing secret without confirmation")
	_ = secretSetCmd.MarkFlagRequired("value")

	secretGetCmd.Flags().BoolVarP(&copyToClip, "copy", "c", false, "Copy value to clipboard instead of printing")
	secretGetCmd.Flags().BoolVar(&copyKeyValue, "copy-kv", false, "Copy KEY=VALUE format to clipboard")

	secretDeleteCmd.Flags().BoolVarP(&forceOverride, "force", "f", false, "Skip confirmation prompt")

	secretExportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (required)")
	_ = secretExportCmd.MarkFlagRequired("output")
}

func runSecretSet(cmd *cobra.Command, args []string) error {
	projectCtx, err := GetProjectContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(args[0])
	value := strings.TrimSpace(secretValue)

	if err := validation.ValidateSecretKey(key); err != nil {
		return err
	}
	if err := validation.ValidateSecretValue(value); err != nil {
		return err
	}

	fmt.Printf("🔐 Setting secret '%s' in project %s...\n", key, projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	projectKey, err := getProjectKey(projectCtx.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get project key: %w", err)
	}

	fmt.Println("🔒 Encrypting secret value...")

	encryptedValue, nonce, err := encryptSecretValue(value, projectKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt secret: %w", err)
	}

	fmt.Println("📡 Uploading encrypted secret to server...")

	c := client.New()
	secret, err := c.SetSecret(
		projectCtx.OrgSlug, projectCtx.ProjectSlug, key,
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

func parseCompositeSlug(compositeSlug string) (string, string, error) {
	parts := strings.Split(compositeSlug, "/")
	const expectedParts = 2
	if len(parts) != expectedParts {
		return "", "", fmt.Errorf(
			"invalid composite slug format: expected 'org-slug/project-slug', got '%s'",
			compositeSlug,
		)
	}
	return parts[0], parts[1], nil
}

func getProjectKey(compositeSlug string, store *storage.Storage) ([]byte, error) {
	orgSlug, projectSlug, err := parseCompositeSlug(compositeSlug)
	if err != nil {
		return nil, err
	}

	c := client.New()
	wrappedKey, err := c.GetWrappedProjectKey(orgSlug, projectSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wrapped project key: %w", err)
	}

	devicePrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get device private key: %w", err)
	}

	projectKey, err := crypto.UnwrapProjectKey(wrappedKey, devicePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap project key: %w", err)
	}

	return projectKey, nil
}

func encryptSecretValue(value string, projectKey []byte) ([]byte, []byte, error) {
	return crypto.EncryptSecretValue(value, projectKey)
}

func runSecretGet(cmd *cobra.Command, args []string) error {
	projectCtx, err := GetProjectContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(args[0])
	if err := validation.ValidateSecretKey(key); err != nil {
		return err
	}

	fmt.Printf("🔍 Getting secret '%s' from project %s...\n", key, projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	projectKey, err := getProjectKey(projectCtx.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get project key: %w", err)
	}

	c := client.New()
	secretData, err := c.GetSecret(projectCtx.OrgSlug, projectCtx.ProjectSlug, key)
	if err != nil {
		return fmt.Errorf("❌ Failed to get secret: %w", err)
	}

	fmt.Println("🔓 Decrypting secret value...")

	decryptedValue, err := decryptSecretValue(secretData.EncryptedValue, secretData.Nonce, projectKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to decrypt secret: %w", err)
	}

	output := map[string]interface{}{
		"key":               secretData.Key,
		"value":             decryptedValue,
		"version":           secretData.Version,
		"project_id":        secretData.ProjectID,
		"updated_at":        secretData.UpdatedAt,
		"created_by_device": secretData.CreatedByDevice.Name,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("❌ Failed to format JSON output: %w", err)
	}

	if copyToClip || copyKeyValue {
		if err := copySecretToClipboard(key, decryptedValue, copyToClip, copyKeyValue); err != nil {
			return err
		}
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func copySecretToClipboard(key, value string, copyToClip, copyKeyValue bool) error {
	if !copyToClip && !copyKeyValue {
		return nil
	}

	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("❌ Failed to initialize clipboard: %w", err)
	}

	if copyToClip {
		clipboard.Write(clipboard.FmtText, []byte(value))
		fmt.Println("📋 Secret value copied to clipboard")
	} else if copyKeyValue {
		keyValue := fmt.Sprintf("%s=%s", key, value)
		clipboard.Write(clipboard.FmtText, []byte(keyValue))
		fmt.Println("📋 KEY=VALUE format copied to clipboard")
	}

	return nil
}

func runSecretList(cmd *cobra.Command, args []string) error {
	projectCtx, err := GetProjectContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔍 Listing secrets in project %s...\n", projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	secrets, err := c.ListSecrets(projectCtx.OrgSlug, projectCtx.ProjectSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to list secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("No secrets found in this project.")
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
	projectCtx, err := GetProjectContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(args[0])
	if err := validation.ValidateSecretKey(key); err != nil {
		return err
	}

	if !forceOverride {
		fmt.Printf("⚠️  Are you sure you want to delete secret '%s' from project %s? (y/N): ",
			key, projectCtx.String())
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	fmt.Printf("🗑️  Deleting secret '%s' from project %s...\n", key, projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	if err := c.DeleteSecret(projectCtx.OrgSlug, projectCtx.ProjectSlug, key); err != nil {
		return fmt.Errorf("❌ Failed to delete secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' deleted successfully!\n", key)
	return nil
}
func decryptSecretValue(encryptedValue, nonce string, projectKey []byte) (string, error) {
	ciphertext, err := crypto.Decode(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}

	nonceBytes, err := crypto.Decode(nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	return crypto.DecryptSecretValue(ciphertext, nonceBytes, projectKey)
}

func runSecretExport(cmd *cobra.Command, args []string) error {
	projectCtx, err := GetProjectContext()
	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	key := strings.TrimSpace(args[0])
	if err := validation.ValidateSecretKey(key); err != nil {
		return err
	}

	fmt.Printf("🔍 Getting secret '%s' from project %s...\n", key, projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	projectKey, err := getProjectKey(projectCtx.String(), store)
	if err != nil {
		return fmt.Errorf("❌ Failed to get project key: %w", err)
	}

	c := client.New()
	secretData, err := c.GetSecret(projectCtx.OrgSlug, projectCtx.ProjectSlug, key)
	if err != nil {
		return fmt.Errorf("❌ Failed to get secret: %w", err)
	}

	fmt.Println("🔓 Decrypting secret value...")

	decryptedValue, err := decryptSecretValue(secretData.EncryptedValue, secretData.Nonce, projectKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to decrypt secret: %w", err)
	}

	exportService := export.NewExportService(forceOverride)
	return exportService.ExportSecret(key, decryptedValue, outputFile)
}
