package project

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/curve25519"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func GetProjectContext(projectPath, org, project string) (*config.ProjectContext, error) {
	projectCtx, err := config.ResolveProjectContext(projectPath, org, project)
	if err == nil {
		return projectCtx, nil
	}

	fmt.Println("❓ Project context is required for this command.")
	fmt.Println("💡 You can specify project using:")
	fmt.Println("   --project-path org/project")
	fmt.Println("   --org org --project project")
	fmt.Println("   Create a .initiat file with org and project")
	fmt.Println("   Or configure defaults with 'initiat config set org <org>' and " +
		"'initiat config set project <project>'")
	fmt.Println()

	projectOptions, fetchErr := getAvailableProjects()
	if fetchErr != nil {
		fmt.Println("⚠️  Could not fetch available projects, please enter manually:")
		promptedProject, promptErr := prompt.PromptProject()
		if promptErr != nil {
			return nil, fmt.Errorf("failed to get project from prompt: %w", promptErr)
		}
		return config.ResolveProjectContext(promptedProject, "", "")
	}

	promptedProject, promptErr := prompt.PromptProjectWithOptions(projectOptions)
	if promptErr != nil {
		return nil, fmt.Errorf("failed to get project from prompt: %w", promptErr)
	}

	return config.ResolveProjectContext(promptedProject, "", "")
}

func getAvailableProjects() ([]prompt.ProjectOption, error) {
	store := storage.New()
	if !store.HasDeviceID() {
		return nil, fmt.Errorf("device not registered")
	}

	c := client.New()
	projects, err := c.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	options := make([]prompt.ProjectOption, len(projects))
	for i, project := range projects {
		compositeSlug := project.CompositeSlug
		if compositeSlug == "" {
			compositeSlug = fmt.Sprintf("%s/%s", project.Organization.Slug, project.Slug)
		}
		options[i] = prompt.ProjectOption{
			Name: project.Name,
			Slug: compositeSlug,
		}
	}

	return options, nil
}

type SetupDetails struct {
	OrgSlug     string
	ProjectSlug string
}

type SetupResult struct {
	Success        bool
	ProjectCreated bool
	KeyInitialized bool
	Message        string
}

func SetupProject(details SetupDetails) (*SetupResult, error) {
	store := storage.New()
	if !store.HasDeviceID() {
		return nil, fmt.Errorf("device not registered. Please run 'initiat device register <name>' first")
	}

	c := client.New()
	project, err := c.GetProjectBySlug(details.OrgSlug, details.ProjectSlug)
	if err != nil {
		return &SetupResult{
			Success:        true,
			ProjectCreated: false,
			KeyInitialized: false,
			Message: fmt.Sprintf(
				"Project '%s/%s' doesn't exist remotely. Please create it at https://www.initiat.dev",
				details.OrgSlug, details.ProjectSlug,
			),
		}, nil
	}

	if !project.KeyInitialized {
		if err := InitializeProjectKey(c, store, project, details.OrgSlug, details.ProjectSlug); err != nil {
			return nil, fmt.Errorf("failed to initialize project key: %w", err)
		}
		return &SetupResult{
			Success:        true,
			ProjectCreated: false,
			KeyInitialized: true,
			Message:        "Project key initialized successfully",
		}, nil
	}

	return &SetupResult{
		Success:        true,
		ProjectCreated: false,
		KeyInitialized: true,
		Message:        "Project key already initialized",
	}, nil
}

func CreateInitiatFile(orgSlug, projectSlug string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	initiatPath := filepath.Join(wd, ".initiat")
	content := fmt.Sprintf("org: %s\nproject: %s\n", orgSlug, projectSlug)

	const initiatFileMode = 0644
	// #nosec G306 - .initiat file contains only org and project names, not sensitive data
	return os.WriteFile(initiatPath, []byte(content), initiatFileMode)
}

func CheckInitiatFileExists() (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current directory: %w", err)
	}

	initiatPath := filepath.Join(wd, ".initiat")
	_, err = os.Stat(initiatPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check .initiat file: %w", err)
	}
	return true, nil
}

func GetDefaultProjectName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Base(wd), nil
}

func InitializeProjectKey(
	c *client.Client, store *storage.Storage, project *types.Project, orgSlug, projectSlug string,
) error {
	projectKey := make([]byte, crypto.ProjectKeySize)
	if _, err := rand.Read(projectKey); err != nil {
		return fmt.Errorf("failed to generate project key: %w", err)
	}

	encryptionPrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get device encryption key: %w", err)
	}

	encryptionPublicKey, err := curve25519.X25519(encryptionPrivateKey, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("failed to derive device public key: %w", err)
	}

	wrappedKeyStr, err := crypto.WrapProjectKey(projectKey, encryptionPublicKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt project key: %w", err)
	}

	wrappedKey, err := crypto.Decode(wrappedKeyStr)
	if err != nil {
		return fmt.Errorf("failed to decode wrapped key: %w", err)
	}

	if err := c.InitializeProjectKey(orgSlug, projectSlug, wrappedKey); err != nil {
		return fmt.Errorf("failed to initialize project key: %w", err)
	}

	return nil
}

func PromptForOrganization() (string, error) {
	defaultOrg := config.GetDefaultOrgSlug()
	if defaultOrg != "" {
		fmt.Printf("🏢 Organization (default: %s): ", defaultOrg)
		orgSlug, err := promptString()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(orgSlug) == "" {
			return defaultOrg, nil
		}
		return strings.TrimSpace(orgSlug), nil
	}

	fmt.Print("🏢 Organization: ")
	orgSlug, err := promptString()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(orgSlug), nil
}

func PromptForProjectName() (string, error) {
	defaultProject, err := GetDefaultProjectName()
	if err != nil {
		return "", err
	}

	fmt.Printf("📦 Project name (default: %s): ", defaultProject)
	projectSlug, err := promptString()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(projectSlug) == "" {
		return defaultProject, nil
	}
	return strings.TrimSpace(projectSlug), nil
}

func promptString() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(input), nil
}
