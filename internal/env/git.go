package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/git"
)

const (
	gitignoreEntry = `# Initiat
.initiat/environments/
.initiat/active`
	GitignorePerms      = 0644
	GitStatusNotRepo    = "not a git repository"
	GitStatusMissing    = "missing"
	GitStatusConfigured = "configured"
)

var gitHandler = git.NewHandler()

func EnsureGitignore() error {
	gitignorePath := ".gitignore"

	content, err := gitHandler.ReadFile(gitignorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read .gitignore: %w", err)
		}
		content = ""
	}

	if strings.Contains(content, "# Initiat") {
		return nil
	}

	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, GitignorePerms)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	if _, err := file.WriteString("\n" + gitignoreEntry + "\n"); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}
	return nil
}

func CheckGitignore() (bool, error) {
	gitignorePath := ".gitignore"

	content, err := gitHandler.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read .gitignore: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == ".initiat/environments/" || line == ".initiat/active" {
			return true, nil
		}
	}

	return false, nil
}

func IsGitRepository() bool {
	return gitHandler.IsGitRepository(".")
}

func GetGitignoreStatus() (string, error) {
	if !IsGitRepository() {
		return GitStatusNotRepo, nil
	}

	hasEntry, err := CheckGitignore()
	if err != nil {
		return "", err
	}

	if hasEntry {
		return GitStatusConfigured, nil
	}

	return GitStatusMissing, nil
}
