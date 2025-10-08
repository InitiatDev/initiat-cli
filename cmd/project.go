package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/project"
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

var projectSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up a new project",
	Long: `Set up a new project by creating a .initiat file and initializing the project.

This command will:
- Create a .initiat file in the current directory
- Prompt for organization (uses default if set)
- Ask for project name (defaults to current folder name)
- Create the project remotely if it doesn't exist
- Initialize the project key

Examples:
  initiat project setup`,
	RunE: runProjectSetup,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectSetupCmd)
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
	proj, err := c.GetProjectBySlug(projectCtx.OrgSlug, projectCtx.ProjectSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to get project info: %w", err)
	}

	if !checkProjectInitStatus(proj) {
		return nil
	}

	return project.InitializeProjectKey(c, store, proj, projectCtx.OrgSlug, projectCtx.ProjectSlug)
}

func checkProjectInitStatus(project *types.Project) bool {
	if project.KeyInitialized {
		fmt.Println("ℹ️ Project key already initialized on server")
		return false
	}
	return true
}

func runProjectSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Setting up a new project...")
	fmt.Println()

	exists, err := project.CheckInitiatFileExists()
	if err != nil {
		return fmt.Errorf("❌ Failed to check for existing .initiat file: %w", err)
	}

	if exists {
		fmt.Println("⚠️  A .initiat file already exists in this directory.")
		fmt.Print("❓ Do you want to overwrite it? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Setup cancelled")
			return nil
		}
	}

	orgSlug, err := project.PromptForOrganization()
	if err != nil {
		return fmt.Errorf("❌ Failed to get organization: %w", err)
	}

	projectSlug, err := project.PromptForProjectName()
	if err != nil {
		return fmt.Errorf("❌ Failed to get project name: %w", err)
	}

	if err := project.CreateInitiatFile(orgSlug, projectSlug); err != nil {
		return fmt.Errorf("❌ Failed to create .initiat file: %w", err)
	}

	fmt.Printf("✅ Created .initiat file with org: %s, project: %s\n", orgSlug, projectSlug)

	details := project.SetupDetails{
		OrgSlug:     orgSlug,
		ProjectSlug: projectSlug,
	}

	result, err := project.SetupProject(details)
	if err != nil {
		return fmt.Errorf("❌ Failed to set up project: %w", err)
	}

	if result.Success {
		if result.KeyInitialized {
			fmt.Println("🔐 Project key initialized successfully!")
		} else {
			fmt.Printf("❌ %s\n", result.Message)
			fmt.Println()
			fmt.Println("💡 To create a new project:")
			fmt.Println("   1. Visit https://www.initiat.dev")
			fmt.Println("   2. Create the project in your organization")
			fmt.Println("   3. Run this setup command again")
			fmt.Println()
			fmt.Println("✅ Local .initiat file has been created.")
			fmt.Println("🔄 Run 'initiat project setup' again after creating the project remotely.")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("🎉 Project setup complete!")
	fmt.Printf("📍 Project: %s/%s\n", orgSlug, projectSlug)
	fmt.Println("📁 Local config: .initiat file created")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Add secrets: initiat secret set API_KEY --value your-secret")
	fmt.Println("  • List secrets: initiat secret list")
	fmt.Println("  • Invite team members: initiat project invite-device")

	return nil
}
