package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/env"
	"github.com/InitiatDev/initiat-cli/internal/project"
	"github.com/InitiatDev/initiat-cli/internal/setup"
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
	Short: "Initialize a project",
	Long: `Initialize a project by creating local .initiat/config.yml and initializing the project key remotely.

This command will:
- Create .initiat/config.yml in the current directory (if in a git repository)
- Prompt for organization (uses default if set)
- Ask for project name (defaults to current folder name)
- Initialize the project key remotely (if not already initialized)

Both operations are idempotent - they will skip if already completed.

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
	Short: "Run project setup script",
	Long: `Run the setup script from .initiat/setup.yml to configure the development environment.

This command will:
- Read and parse .initiat/setup.yml
- Validate the setup configuration
- Execute the setup script (install tools, runtimes, databases, etc.)

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
	fmt.Println("Fetching projects...")

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
		fmt.Println("\nInitialize keys for projects marked \"No\" using:")
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

	fmt.Printf("Initializing project \"%s\"...\n", projectCtx.String())

	store := storage.New()
	if !store.HasDeviceID() {
		return fmt.Errorf("❌ Device not registered. Please run 'initiat device register <name>' first")
	}

	orgSlug := projectCtx.OrgSlug
	projectSlug := projectCtx.ProjectSlug

	if err := ensureInitiatFileExists(orgSlug, projectSlug); err != nil {
		return fmt.Errorf("❌ Failed to create .initiat/config.yml: %w", err)
	}

	c := client.New()
	proj, err := c.GetProjectBySlug(orgSlug, projectSlug)
	if err != nil {
		return fmt.Errorf("❌ Failed to get project info: %w", err)
	}

	if !checkProjectInitStatus(proj) {
		return nil
	}

	if err := project.InitializeProjectKey(c, store, proj, orgSlug, projectSlug); err != nil {
		return fmt.Errorf("❌ Failed to initialize project key: %w", err)
	}

	fmt.Printf("✅ Project \"%s\" initialized successfully!\n", projectCtx.String())
	return nil
}

func ensureInitiatFileExists(orgSlug, projectSlug string) error {
	exists, err := project.CheckInitiatFileExists()
	if err != nil {
		return err
	}

	if exists {
		fmt.Println("✅ .initiat/config.yml already exists")
		return nil
	}

	if !env.IsGitRepository() {
		fmt.Println("⚠️  Not in a git repository - skipping .initiat/config.yml creation")
		fmt.Println("💡 Create a git repository first, or manually create .initiat/config.yml")
		return nil
	}

	fmt.Println("Creating .initiat/config.yml...")
	if err := project.CreateInitiatFile(orgSlug, projectSlug); err != nil {
		return err
	}

	fmt.Printf("✅ Created .initiat/config.yml with org: %s, project: %s\n", orgSlug, projectSlug)
	return nil
}

func checkProjectInitStatus(project *types.Project) bool {
	if project.KeyInitialized {
		fmt.Println("Project key already initialized on server")
		return false
	}
	return true
}

func runProjectSetup(cmd *cobra.Command, args []string) error {
	setupFile := ".initiat/setup.yml"

	projectCtx, err := GetProjectContext()
	if err != nil {
		return fmt.Errorf("❌ Failed to get project context: %w", err)
	}

	fmt.Printf("📋 Loading setup script from %s...\n", setupFile)
	config, err := setup.ParseFile(setupFile)
	if err != nil {
		return fmt.Errorf("❌ Failed to parse setup file: %w", err)
	}

	fmt.Println("🔍 Validating setup configuration...")
	if err := setup.Validate(config); err != nil {
		fmt.Println("❌ Validation failed:")
		if validationErrs, ok := err.(setup.ValidationErrors); ok {
			for _, validationErr := range validationErrs {
				fmt.Printf("  - %s\n", validationErr.Error())
			}
		} else {
			fmt.Printf("  - %s\n", err.Error())
		}
		return fmt.Errorf("validation failed")
	}

	fmt.Println("✅ Setup configuration is valid")

	fmt.Println("🔧 Creating execution context...")
	fmt.Println("📝 Generating execution plan...")

	runner := setup.NewSetupRunner(projectCtx)
	if err := runner.Run(config); err != nil {
		if errors.Is(err, setup.ErrNoCommandsToExecute) {
			fmt.Println("ℹ️  No commands to execute (all steps skipped due to conditions)")
			return nil
		}
		return fmt.Errorf("❌ Setup execution failed: %w", err)
	}

	return nil
}
