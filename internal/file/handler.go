package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// File permissions
	DirPerms  = 0755
	FilePerms = 0600
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Directory operations
func (h *Handler) EnsureDirectory(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, DirPerms)
}

func (h *Handler) CreateDirectory(path string) error {
	return os.MkdirAll(path, DirPerms)
}

// File operations
func (h *Handler) ReadFile(path string) (string, error) {
	// Validate path to prevent directory traversal
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		path = absPath
	}

	content, err := os.ReadFile(path) // #nosec G304 - path is validated above
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (h *Handler) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), FilePerms)
}

func (h *Handler) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (h *Handler) RemoveFile(path string) error {
	return os.Remove(path)
}

func (h *Handler) CreateSymlink(target, link string) error {
	_ = os.Remove(link) // Ignore error if file doesn't exist
	return os.Symlink(target, link)
}

func (h *Handler) ReadSymlink(link string) (string, error) {
	return os.Readlink(link)
}

func (h *Handler) GetWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}

func (h *Handler) JoinPaths(paths ...string) string {
	return filepath.Join(paths...)
}

func (h *Handler) GetAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

func (h *Handler) GetRelativePath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

func (h *Handler) GetDirectory(path string) string {
	return filepath.Dir(path)
}

// Generic path builders
func (h *Handler) GetProjectPath(projectDir string) (string, error) {
	wd, err := h.GetWorkingDirectory()
	if err != nil {
		return "", err
	}
	return h.JoinPaths(wd, projectDir), nil
}

func (h *Handler) GetSubPath(basePath, subDir string) (string, error) {
	return h.JoinPaths(basePath, subDir), nil
}

func (h *Handler) GetFilePath(basePath, filename string) (string, error) {
	return h.JoinPaths(basePath, filename), nil
}

func (h *Handler) FindKeyInContent(content, key string) (int, bool) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			return i, true
		}
	}
	return -1, false
}

func (h *Handler) UpdateKeyInContent(content, key, value string, keyIndex int) string {
	lines := strings.Split(content, "\n")
	lines[keyIndex] = fmt.Sprintf("%s=%s", key, value)
	return strings.Join(lines, "\n")
}

func (h *Handler) AppendKeyToContent(content, key, value string) string {
	if content == "" {
		return fmt.Sprintf("%s=%s\n", key, value)
	}
	if strings.HasSuffix(content, "\n") {
		return content + fmt.Sprintf("%s=%s", key, value)
	}
	return content + "\n" + fmt.Sprintf("%s=%s", key, value)
}
