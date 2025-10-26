package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	gitignoreEntry = `# Initiat
.initiat/environments/*/secrets.env
.initiat/active`
	GitignorePerms = 0644
)

func EnsureGitignore() error {
	gitignorePath := ".gitignore"

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read .gitignore: %w", err)
		}
		content = []byte{}
	}

	if strings.Contains(string(content), "# Initiat") {
		return nil
	}

	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, GitignorePerms)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
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

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read .gitignore: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == ".initiat/environments/*/secrets.env" || line == ".initiat/active" {
			return true, nil
		}
	}

	return false, nil
}

func IsGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func GetGitignoreStatus() (string, error) {
	if !IsGitRepository() {
		return "not a git repository", nil
	}

	hasEntry, err := CheckGitignore()
	if err != nil {
		return "", err
	}

	if hasEntry {
		return "configured", nil
	}

	return "missing", nil
}
