package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptEmail() (string, error) {
	fmt.Print("Email: ")
	reader := bufio.NewReader(os.Stdin)
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}
	return email, nil
}

func PromptPassword() (string, error) {
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	password := string(passwordBytes)
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	return password, nil
}

// PromptWorkspace prompts the user for a workspace path
func PromptWorkspace() (string, error) {
	fmt.Print("Workspace (org/workspace): ")
	reader := bufio.NewReader(os.Stdin)
	workspace, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read workspace: %w", err)
	}
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return "", fmt.Errorf("workspace cannot be empty")
	}
	return workspace, nil
}

// PromptWorkspaceWithOptions prompts the user to select from available workspaces or enter a custom one
func PromptWorkspaceWithOptions(workspaces []WorkspaceOption) (string, error) {
	if len(workspaces) == 0 {
		return PromptWorkspace()
	}

	fmt.Println("Available workspaces:")
	for i, workspace := range workspaces {
		fmt.Printf("  %d. %s (%s)\n", i+1, workspace.Name, workspace.Slug)
	}
	fmt.Println("  0. Enter custom workspace")
	fmt.Println()

	fmt.Print("Select workspace (0 for custom): ")
	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}
	choice = strings.TrimSpace(choice)

	if choice == "0" {
		return PromptWorkspace()
	}

	// Parse the choice as a number
	var index int
	if _, err := fmt.Sscanf(choice, "%d", &index); err != nil {
		return "", fmt.Errorf("invalid choice: %s", choice)
	}

	if index < 1 || index > len(workspaces) {
		return "", fmt.Errorf("invalid choice: %d (must be between 1 and %d)", index, len(workspaces))
	}

	return workspaces[index-1].Slug, nil
}

// WorkspaceOption represents a workspace option for selection
type WorkspaceOption struct {
	Name string
	Slug string
}
