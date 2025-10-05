package export

import (
	"fmt"
	"path/filepath"
)

type ExportService struct {
	fileHandler   *FileHandler
	gitHandler    *GitHandler
	promptHandler *PromptHandler
	forceOverride bool
}

func NewExportService(forceOverride bool) *ExportService {
	return &ExportService{
		fileHandler:   NewFileHandler(),
		gitHandler:    NewGitHandler(),
		promptHandler: NewPromptHandler(),
		forceOverride: forceOverride,
	}
}

func (e *ExportService) ExportSecret(key, value, filePath string) error {
	if err := e.fileHandler.EnsureDirectory(filePath); err != nil {
		return fmt.Errorf("❌ Failed to create directory: %w", err)
	}

	if err := e.handleFileExport(key, value, filePath); err != nil {
		return err
	}

	if err := e.handleGitIgnore(filePath); err != nil {
		return err
	}

	fmt.Printf("✅ Secret '%s' exported to %s\n", key, filePath)
	return nil
}

func (e *ExportService) handleFileExport(key, value, filePath string) error {
	var content string
	var err error

	if e.fileHandler.FileExists(filePath) {
		content, err = e.fileHandler.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("❌ Failed to read existing file: %w", err)
		}
	}

	keyIndex, keyExists := e.fileHandler.FindKeyInContent(content, key)

	if keyExists {
		if !e.forceOverride && !e.promptHandler.ConfirmOverwrite(key) {
			fmt.Println("❌ Export cancelled.")
			return nil
		}
		content = e.fileHandler.UpdateKeyInContent(content, key, value, keyIndex)
	} else {
		content = e.fileHandler.AppendKeyToContent(content, key, value)
	}

	if err := e.fileHandler.WriteFile(filePath, content); err != nil {
		return fmt.Errorf("❌ Failed to write file: %w", err)
	}

	return nil
}

func (e *ExportService) handleGitIgnore(filePath string) error {
	gitRoot, found := e.gitHandler.FindGitRoot(filePath)
	if !found {
		return nil
	}

	gitignoreContent, err := e.gitHandler.ReadGitignore(gitRoot)
	if err != nil {
		return fmt.Errorf("❌ Failed to read .gitignore: %w", err)
	}

	if gitignoreContent == "" {
		return nil
	}

	if !e.gitHandler.IsFileIgnored(gitignoreContent, filePath, gitRoot) {
		relativePath, err := filepath.Rel(gitRoot, filePath)
		if err != nil {
			return fmt.Errorf("❌ Failed to get relative path: %w", err)
		}

		if e.promptHandler.ConfirmGitignore(relativePath) {
			if err := e.gitHandler.AddToGitignore(gitRoot, filePath); err != nil {
				return fmt.Errorf("❌ Failed to update .gitignore: %w", err)
			}
			fmt.Printf("✅ Added '%s' to .gitignore\n", relativePath)
		}
	}

	return nil
}
