package workspace

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
	"github.com/InitiatDev/initiat-cli/internal/storage"
)

func GetWorkspaceContext(workspacePath, org, workspace string) (*config.WorkspaceContext, error) {
	workspaceCtx, err := config.ResolveWorkspaceContext(workspacePath, org, workspace)
	if err == nil {
		return workspaceCtx, nil
	}

	fmt.Println("❓ Workspace context is required for this command.")
	fmt.Println("💡 You can specify workspace using:")
	fmt.Println("   --workspace-path org/workspace")
	fmt.Println("   --org org --workspace workspace")
	fmt.Println("   Or configure defaults with 'initiat config set org <org>' and " +
		"'initiat config set workspace <workspace>'")
	fmt.Println()

	workspaceOptions, fetchErr := getAvailableWorkspaces()
	if fetchErr != nil {
		fmt.Println("⚠️  Could not fetch available workspaces, please enter manually:")
		promptedWorkspace, promptErr := prompt.PromptWorkspace()
		if promptErr != nil {
			return nil, fmt.Errorf("failed to get workspace from prompt: %w", promptErr)
		}
		return config.ResolveWorkspaceContext(promptedWorkspace, "", "")
	}

	promptedWorkspace, promptErr := prompt.PromptWorkspaceWithOptions(workspaceOptions)
	if promptErr != nil {
		return nil, fmt.Errorf("failed to get workspace from prompt: %w", promptErr)
	}

	return config.ResolveWorkspaceContext(promptedWorkspace, "", "")
}

func getAvailableWorkspaces() ([]prompt.WorkspaceOption, error) {
	store := storage.New()
	if !store.HasDeviceID() {
		return nil, fmt.Errorf("device not registered")
	}

	c := client.New()
	workspaces, err := c.ListWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspaces: %w", err)
	}

	options := make([]prompt.WorkspaceOption, len(workspaces))
	for i, workspace := range workspaces {
		compositeSlug := workspace.CompositeSlug
		if compositeSlug == "" {
			compositeSlug = fmt.Sprintf("%s/%s", workspace.Organization.Slug, workspace.Slug)
		}
		options[i] = prompt.WorkspaceOption{
			Name: workspace.Name,
			Slug: compositeSlug,
		}
	}

	return options, nil
}
