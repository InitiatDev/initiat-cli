package export

import (
	"fmt"
	"strings"
)

type PromptHandler struct{}

func NewPromptHandler() *PromptHandler {
	return &PromptHandler{}
}

func (p *PromptHandler) ConfirmOverwrite(key string) bool {
	fmt.Printf("⚠️  Key '%s' already exists in file. Overwrite? (y/N): ", key)
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func (p *PromptHandler) ConfirmGitignore(filePath string) bool {
	fmt.Printf("⚠️  File '%s' is not in .gitignore. Add it? (y/N): ", filePath)
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
