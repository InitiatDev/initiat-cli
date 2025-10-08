package project

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
	"github.com/InitiatDev/initiat-cli/internal/storage"
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
