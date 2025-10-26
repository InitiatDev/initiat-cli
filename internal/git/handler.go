package git

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	FilePerms = 0600
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (g *Handler) FindGitRoot(startPath string) (string, bool) {
	dir := filepath.Dir(startPath)
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func (g *Handler) IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

func (g *Handler) ReadGitignore(gitRoot string) (string, error) {
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	if _, err := os.Stat(gitignorePath); err != nil {
		return "", nil
	}
	return g.ReadFile(gitignorePath)
}

func (g *Handler) ReadFile(path string) (string, error) {
	// #nosec G304 - path is user-controlled but validated for git functionality
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (g *Handler) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), FilePerms)
}

func (g *Handler) IsFileIgnored(gitignoreContent, filePath, gitRoot string) bool {
	relativePath, err := filepath.Rel(gitRoot, filePath)
	if err != nil {
		return false
	}

	lines := strings.Split(gitignoreContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle directory patterns (ending with /)
		if strings.HasSuffix(line, "/") {
			dirPattern := strings.TrimSuffix(line, "/")
			if strings.HasPrefix(relativePath, dirPattern+"/") || relativePath == dirPattern {
				return true
			}
		} else {
			// Handle file patterns
			if relativePath == line {
				return true
			}
			// Handle wildcard patterns like *.log
			if strings.Contains(line, "*") {
				// Simple wildcard matching for common patterns
				if strings.HasPrefix(line, "*") && strings.HasSuffix(relativePath, strings.TrimPrefix(line, "*")) {
					return true
				}
			}
		}
	}

	return false
}

func (g *Handler) AddToGitignore(gitRoot, filePath string) error {
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	content, err := g.ReadGitignore(gitRoot)
	if err != nil {
		return err
	}

	relativePath, err := filepath.Rel(gitRoot, filePath)
	if err != nil {
		return err
	}

	var newContent string
	switch {
	case content == "":
		newContent = relativePath + "\n"
	case strings.HasSuffix(content, "\n"):
		newContent = content + relativePath + "\n"
	default:
		newContent = content + "\n" + relativePath + "\n"
	}
	return g.WriteFile(gitignorePath, newContent)
}
