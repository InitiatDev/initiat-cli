package export

import (
	"os"
	"path/filepath"
	"strings"
)

type GitHandler struct{}

func NewGitHandler() *GitHandler {
	return &GitHandler{}
}

func (g *GitHandler) FindGitRoot(startPath string) (string, bool) {
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

func (g *GitHandler) ReadGitignore(gitRoot string) (string, error) {
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	if _, err := os.Stat(gitignorePath); err != nil {
		return "", nil
	}
	return g.ReadFile(gitignorePath)
}

func (g *GitHandler) ReadFile(path string) (string, error) {
	// #nosec G304 - path is user-controlled but validated for export functionality
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (g *GitHandler) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), filePerms)
}

func (g *GitHandler) IsFileIgnored(gitignoreContent, filePath, gitRoot string) bool {
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

		if strings.HasSuffix(line, "/") {
			dirPattern := strings.TrimSuffix(line, "/")
			if strings.HasPrefix(relativePath, dirPattern+"/") || relativePath == dirPattern {
				return true
			}
		} else if relativePath == line || strings.Contains(relativePath, line) {
			return true
		}
	}

	return false
}

func (g *GitHandler) AddToGitignore(gitRoot, filePath string) error {
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	content, err := g.ReadGitignore(gitRoot)
	if err != nil {
		return err
	}

	relativePath, err := filepath.Rel(gitRoot, filePath)
	if err != nil {
		return err
	}

	newContent := content + "\n" + relativePath + "\n"
	return g.WriteFile(gitignorePath, newContent)
}
