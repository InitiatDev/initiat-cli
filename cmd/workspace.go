package cmd

import (
	"crypto/rand"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/config"
	"github.com/DylanBlakemore/initiat-cli/internal/encoding"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
	"github.com/DylanBlakemore/initiat-cli/internal/table"
	"github.com/DylanBlakemore/initiat-cli/internal/types"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces and workspace keys",
	Long:  `Manage workspaces and workspace keys for secure secret storage.`,
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `List all workspaces and their key initialization status.`,
	RunE:  runWorkspaceList,
}

var workspaceInitCmd = &cobra.Command{
	Use:   "init [workspace-path]",
	Short: "Initialize workspace key",
	Long: `Initialize a new workspace key for secure secret storage.

Examples:
  initiat workspace init acme-corp/production
  initiat workspace init acme-corp/production  # Using positional argument
  initiat workspace init --org acme-corp --workspace production
  initiat workspace init -o acme-corp -w production
  initiat workspace init --workspace production  # Uses default org
  initiat workspace init -w production
  initiat workspace init -W prod  # Using alias`,
	RunE: runWorkspaceInit,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceInitCmd)
}

func runWorkspaceList(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Fetching workspaces...")

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	workspaces, err := c.ListWorkspaces()
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found. Create one at https://www.initiat.dev")
		return nil
	}

	t := table.New()
	t.SetHeaders("Name", "Composite Slug", "Key Initialized", "Role")

	for _, workspace := range workspaces {
		keyStatus := "❌ No"
		if workspace.KeyInitialized {
			keyStatus = "✅ Yes"
		}

		compositeSlug := workspace.CompositeSlug
		if compositeSlug == "" {
			compositeSlug = fmt.Sprintf("%s/%s", workspace.Organization.Slug, workspace.Slug)
		}

		t.AddRow(workspace.Name, compositeSlug, keyStatus, workspace.Role)
	}

	err = t.Render()
	if err != nil {
		return err
	}

	hasUninitialized := false
	for _, workspace := range workspaces {
		if !workspace.KeyInitialized {
			hasUninitialized = true
			break
		}
	}

	if hasUninitialized {
		fmt.Println("\n💡 Initialize keys for workspaces marked \"No\" using:")
		fmt.Println("   initiat workspace init <org-slug/workspace-slug>")
	}

	return nil
}

func runWorkspaceInit(cmd *cobra.Command, args []string) error {
	var workspaceCtx *config.WorkspaceContext
	var err error

	// Check for positional argument first, then fall back to flags
	if len(args) > 0 {
		workspaceCtx, err = config.ResolveWorkspaceContext(args[0], "", "")
	} else {
		workspaceCtx, err = GetWorkspaceContext()
	}

	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔐 Initializing workspace key for \"%s\"...\n", workspaceCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	workspace, err := c.GetWorkspaceBySlug(workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to get workspace info: %w", err)
	}

	if !checkWorkspaceInitStatus(workspace) {
		return nil
	}

	return initializeWorkspaceKey(c, store, workspace, workspaceCtx)
}

func checkWorkspaceInitStatus(workspace *types.Workspace) bool {
	if workspace.KeyInitialized {
		fmt.Println("ℹ️ Workspace key already initialized on server")
		return false
	}
	return true
}

func initializeWorkspaceKey(
	c *client.Client, store *storage.Storage, _ *types.Workspace, workspaceCtx *config.WorkspaceContext,
) error {
	fmt.Println("⚡ Generating secure 256-bit workspace key...")
	workspaceKey := make([]byte, encoding.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		return fmt.Errorf("❌ Failed to generate workspace key: %w", err)
	}

	encryptionPrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return fmt.Errorf("❌ Failed to get device encryption key: %w", err)
	}

	encryptionPublicKey, err := curve25519.X25519(encryptionPrivateKey, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("❌ Failed to derive device public key: %w", err)
	}

	fmt.Println("🔒 Encrypting workspace key with your device's X25519 key...")
	wrappedKeyStr, err := encoding.WrapWorkspaceKey(workspaceKey, encryptionPublicKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt workspace key: %w", err)
	}

	wrappedKey, err := encoding.Decode(wrappedKeyStr)
	if err != nil {
		return fmt.Errorf("❌ Failed to decode wrapped key: %w", err)
	}

	fmt.Println("📡 Uploading encrypted key to server...")
	if err := c.InitializeWorkspaceKey(workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug, wrappedKey); err != nil {
		return fmt.Errorf("❌ Failed to initialize workspace key: %w", err)
	}

	printSuccessMessage()
	return nil
}

func printSuccessMessage() {
	fmt.Println("✅ Workspace key initialized successfully!")
	fmt.Println("🎯 You can now store and retrieve secrets in this workspace.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Add secrets: initiat secret set API_KEY --value your-secret")
	fmt.Println("  • List secrets: initiat secret list")
	fmt.Println("  • Invite devices: initiat workspace invite-device")
}
