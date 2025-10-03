package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/config"
	"github.com/DylanBlakemore/initiat-cli/internal/encoding"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
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
	Use:   "init",
	Short: "Initialize workspace key",
	Long: `Initialize a new workspace key for secure secret storage.

Examples:
  initiat workspace init --workspace-path acme-corp/production
  initiat workspace init -W acme-corp/production
  initiat workspace init --org acme-corp --workspace production
  initiat workspace init -o acme-corp -w production
  initiat workspace init --workspace production  # Uses default org
  initiat workspace init -w production
  initiat workspace init -W prod  # Using alias`,
	RunE: runWorkspaceInit,
}

var (
	forceWorkspaceInit bool
)

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceInitCmd)
	workspaceInitCmd.Flags().BoolVarP(
		&forceWorkspaceInit,
		"force",
		"f",
		false,
		"Force re-initialization even if local key exists",
	)
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Name\tComposite Slug\tKey Initialized\tRole")
	fmt.Fprintln(w, "────\t──────────────\t───────────────\t────")

	for _, workspace := range workspaces {
		keyStatus := "❌ No"
		if workspace.KeyInitialized {
			keyStatus = "✅ Yes"
		}

		compositeSlug := workspace.CompositeSlug
		if compositeSlug == "" {
			compositeSlug = fmt.Sprintf("%s/%s", workspace.Organization.Slug, workspace.Slug)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			workspace.Name,
			compositeSlug,
			keyStatus,
			workspace.Role)
	}
	_ = w.Flush()

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
	workspaceCtx, err := GetWorkspaceContext()
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

	shouldContinue, err := checkWorkspaceInitStatus(workspace, store, workspaceCtx.String())
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	if err := handleForceFlag(store, workspaceCtx.String()); err != nil {
		return err
	}

	return initializeWorkspaceKey(c, store, workspace, workspaceCtx)
}

func checkWorkspaceInitStatus(workspace *types.Workspace, store *storage.Storage, compositeSlug string) (bool, error) {
	if workspace.KeyInitialized {
		if store.HasWorkspaceKey(compositeSlug) {
			fmt.Println("ℹ️ Workspace key already exists locally and is initialized on server")
			return false, nil
		}
		return false, fmt.Errorf(
			"ℹ️ Workspace key already initialized on server but not found locally. Contact support for key recovery")
	}

	if store.HasWorkspaceKey(compositeSlug) && !forceWorkspaceInit {
		fmt.Println("⚠️  Local workspace key exists but server workspace is not initialized.")
		fmt.Println("   This usually happens when a workspace was deleted and recreated with the same name.")
		fmt.Println("   Use --force to generate a new key and reinitialize.")
		return false, nil
	}

	return true, nil
}

func handleForceFlag(store *storage.Storage, compositeSlug string) error {
	if forceWorkspaceInit && store.HasWorkspaceKey(compositeSlug) {
		fmt.Println("🔄 Force flag detected - removing old local workspace key...")
		if err := store.DeleteWorkspaceKey(compositeSlug); err != nil {
			fmt.Printf("⚠️  Warning: Failed to delete old workspace key: %v\n", err)
		}
	}
	return nil
}

func initializeWorkspaceKey(
	c *client.Client, store *storage.Storage, _ *types.Workspace, workspaceCtx *config.WorkspaceContext,
) error {
	fmt.Println("⚡ Generating secure 256-bit workspace key...")
	workspaceKey := make([]byte, encoding.WorkspaceKeySize)
	if _, err := rand.Read(workspaceKey); err != nil {
		return fmt.Errorf("❌ Failed to generate workspace key: %w", err)
	}

	fmt.Println("🔒 Encrypting with your device's X25519 key...")
	wrappedKey, err := wrapWorkspaceKey(workspaceKey, store)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt workspace key: %w", err)
	}

	fmt.Println("📡 Uploading encrypted key to server...")
	if err := c.InitializeWorkspaceKey(workspaceCtx.OrgSlug, workspaceCtx.WorkspaceSlug, wrappedKey); err != nil {
		return fmt.Errorf("❌ Failed to initialize workspace key: %w", err)
	}

	if err := store.StoreWorkspaceKey(workspaceCtx.String(), workspaceKey); err != nil {
		return fmt.Errorf("❌ Failed to store workspace key locally: %w", err)
	}

	printSuccessMessage()
	return nil
}

func printSuccessMessage() {
	fmt.Println("✅ Workspace key initialized successfully!")
	fmt.Println("🎯 You can now store and retrieve secrets in this workspace.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Add secrets: initiat secrets add API_KEY=your-secret")
	fmt.Println("  • List secrets: initiat secrets list")
	fmt.Println("  • Invite devices: initiat workspace invite-device")
}

func wrapWorkspaceKey(workspaceKey []byte, store *storage.Storage) ([]byte, error) {
	encryptionPrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption private key: %w", err)
	}

	ephemeralPrivate := make([]byte, encoding.X25519PrivateKeySize)
	if _, err := rand.Read(ephemeralPrivate); err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral private key: %w", err)
	}

	ephemeralPublic, err := curve25519.X25519(ephemeralPrivate, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral public key: %w", err)
	}

	sharedSecret, err := curve25519.X25519(encryptionPrivateKey, ephemeralPublic)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shared secret: %w", err)
	}

	hkdf := hkdf.New(sha256.New, sharedSecret, []byte("initiat.wrap"), []byte("workspace"))
	encryptionKey := make([]byte, encoding.WorkspaceKeySize)
	if _, err := hkdf.Read(encryptionKey); err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}

	cipher, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	const chacha20NonceSize = 12             // ChaCha20-Poly1305 nonce size
	nonce := make([]byte, chacha20NonceSize) // ChaCha20-Poly1305 nonce for workspace key wrapping
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := cipher.Seal(nil, nonce, workspaceKey, nil) // #nosec G407 - nonce is randomly generated above

	wrapped := make([]byte, 0, 32+12+len(ciphertext))
	wrapped = append(wrapped, ephemeralPublic...)
	wrapped = append(wrapped, nonce...)
	wrapped = append(wrapped, ciphertext...)

	return wrapped, nil
}
