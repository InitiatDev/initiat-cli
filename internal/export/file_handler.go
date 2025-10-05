package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirPerms  = 0o755
	filePerms = 0o600
)

type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (f *FileHandler) EnsureDirectory(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, dirPerms)
}

func (f *FileHandler) ReadFile(path string) (string, error) {
	// #nosec G304 - path is user-controlled but validated for export functionality
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (f *FileHandler) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), filePerms)
}

func (f *FileHandler) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (f *FileHandler) FindKeyInContent(content, key string) (int, bool) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			return i, true
		}
	}
	return -1, false
}

func (f *FileHandler) UpdateKeyInContent(content, key, value string, keyIndex int) string {
	lines := strings.Split(content, "\n")
	lines[keyIndex] = fmt.Sprintf("%s=%s", key, value)
	return strings.Join(lines, "\n")
}

func (f *FileHandler) AppendKeyToContent(content, key, value string) string {
	if content == "" {
		return fmt.Sprintf("%s=%s\n", key, value)
	}
	if strings.HasSuffix(content, "\n") {
		return content + fmt.Sprintf("%s=%s", key, value)
	}
	return content + "\n" + fmt.Sprintf("%s=%s", key, value)
}
