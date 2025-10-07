package cmd

import (
	"crypto/rand"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/table"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects and project keys",
	Long:  `Manage projects and project keys for secure secret storage.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `List all projects and their key initialization status.`,
	RunE:  runProjectList,
}

var projectInitCmd = &cobra.Command{
	Use:   "init [project-path]",
	Short: "Initialize project key",
	Long: `Initialize a new project key for secure secret storage.

Examples:
  initiat project init acme-corp/production
  initiat project init acme-corp/production  # Using positional argument
  initiat project init --org acme-corp --project production
  initiat project init -o acme-corp -p production
  initiat project init --project production  # Uses default org
  initiat project init -p production
  initiat project init -P prod  # Using alias`,
	RunE: runProjectInit,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectInitCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Fetching projects...")

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	projects, err := c.ListProjects()
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found. Create one at https://www.initiat.dev")
		return nil
	}

	t := table.New()
	t.SetHeaders("Name", "Composite Slug", "Key Initialized", "Role")

	for _, project := range projects {
		keyStatus := "❌ No"
		if project.KeyInitialized {
			keyStatus = "✅ Yes"
		}

		compositeSlug := project.CompositeSlug
		if compositeSlug == "" {
			compositeSlug = fmt.Sprintf("%s/%s", project.Organization.Slug, project.Slug)
		}

		t.AddRow(project.Name, compositeSlug, keyStatus, project.Role)
	}

	err = t.Render()
	if err != nil {
		return err
	}

	hasUninitialized := false
	for _, project := range projects {
		if !project.KeyInitialized {
			hasUninitialized = true
			break
		}
	}

	if hasUninitialized {
		fmt.Println("\n💡 Initialize keys for projects marked \"No\" using:")
		fmt.Println("   initiat project init <org-slug/project-slug>")
	}

	return nil
}

func runProjectInit(cmd *cobra.Command, args []string) error {
	var projectCtx *config.ProjectContext
	var err error

	// Check for positional argument first, then fall back to flags
	if len(args) > 0 {
		projectCtx, err = config.ResolveProjectContext(args[0], "", "")
	} else {
		projectCtx, err = GetProjectContext()
	}

	if err != nil {
		return fmt.Errorf("❌ %w", err)
	}

	fmt.Printf("🔐 Initializing project key for \"%s\"...\n", projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	project, err := c.GetProjectBySlug(projectCtx.OrgSlug, projectCtx.ProjectSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to get project info: %w", err)
	}

	if !checkProjectInitStatus(project) {
		return nil
	}

	return initializeProjectKey(c, store, project, projectCtx)
}

func checkProjectInitStatus(project *types.Project) bool {
	if project.KeyInitialized {
		fmt.Println("ℹ️ Project key already initialized on server")
		return false
	}
	return true
}

func initializeProjectKey(
	c *client.Client, store *storage.Storage, _ *types.Project, projectCtx *config.ProjectContext,
) error {
	fmt.Println("⚡ Generating secure 256-bit project key...")
	projectKey := make([]byte, crypto.ProjectKeySize)
	if _, err := rand.Read(projectKey); err != nil {
		return fmt.Errorf("❌ Failed to generate project key: %w", err)
	}

	encryptionPrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return fmt.Errorf("❌ Failed to get device encryption key: %w", err)
	}

	encryptionPublicKey, err := curve25519.X25519(encryptionPrivateKey, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("❌ Failed to derive device public key: %w", err)
	}

	fmt.Println("🔒 Encrypting project key with your device's X25519 key...")
	wrappedKeyStr, err := crypto.WrapProjectKey(projectKey, encryptionPublicKey)
	if err != nil {
		return fmt.Errorf("❌ Failed to encrypt project key: %w", err)
	}

	wrappedKey, err := crypto.Decode(wrappedKeyStr)
	if err != nil {
		return fmt.Errorf("❌ Failed to decode wrapped key: %w", err)
	}

	fmt.Println("📡 Uploading encrypted key to server...")
	if err := c.InitializeProjectKey(projectCtx.OrgSlug, projectCtx.ProjectSlug, wrappedKey); err != nil {
		return fmt.Errorf("❌ Failed to initialize project key: %w", err)
	}

	printSuccessMessage()
	return nil
}

func printSuccessMessage() {
	fmt.Println("✅ Project key initialized successfully!")
	fmt.Println("🎯 You can now store and retrieve secrets in this project.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Add secrets: initiat secret set API_KEY --value your-secret")
	fmt.Println("  • List secrets: initiat secret list")
	fmt.Println("  • Invite devices: initiat project invite-device")
}
