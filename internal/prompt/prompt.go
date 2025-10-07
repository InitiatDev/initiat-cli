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

// PromptProject prompts the user for a project path
func PromptProject() (string, error) {
	fmt.Print("Project (org/project): ")
	reader := bufio.NewReader(os.Stdin)
	project, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read project: %w", err)
	}
	project = strings.TrimSpace(project)
	if project == "" {
		return "", fmt.Errorf("project cannot be empty")
	}
	return project, nil
}

// PromptProjectWithOptions prompts the user to select from available projects or enter a custom one
func PromptProjectWithOptions(projects []ProjectOption) (string, error) {
	if len(projects) == 0 {
		return PromptProject()
	}

	fmt.Println("Available projects:")
	for i, project := range projects {
		fmt.Printf("  %d. %s (%s)\n", i+1, project.Name, project.Slug)
	}
	fmt.Println("  0. Enter custom project")
	fmt.Println()

	fmt.Print("Select project (0 for custom): ")
	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}
	choice = strings.TrimSpace(choice)

	if choice == "0" {
		return PromptProject()
	}

	// Parse the choice as a number
	var index int
	if _, err := fmt.Sscanf(choice, "%d", &index); err != nil {
		return "", fmt.Errorf("invalid choice: %s", choice)
	}

	if index < 1 || index > len(projects) {
		return "", fmt.Errorf("invalid choice: %d (must be between 1 and %d)", index, len(projects))
	}

	return projects[index-1].Slug, nil
}

// ProjectOption represents a project option for selection
type ProjectOption struct {
	Name string
	Slug string
}
