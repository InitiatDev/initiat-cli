package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/file"
)

const (
	InitiatDir          = ".initiat"
	EnvironmentsDir     = "environments"
	ActiveFile          = "active"
	EnvrcFile           = ".envrc"
	WindowsOS           = "windows"
	GitignoreConfigured = "configured"
)

var fileHandler = file.NewHandler()

func GetInitiatPath() (string, error) {
	return fileHandler.GetProjectPath(InitiatDir)
}

func GetEnvironmentsPath() (string, error) {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return "", err
	}
	return fileHandler.GetSubPath(initiatPath, EnvironmentsDir)
}

func GetActivePath() (string, error) {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return "", err
	}
	return fileHandler.GetFilePath(initiatPath, ActiveFile)
}

func GetEnvironmentPath(slug string) (string, error) {
	envsPath, err := GetEnvironmentsPath()
	if err != nil {
		return "", err
	}
	return fileHandler.GetSubPath(envsPath, slug)
}

func CreateInitiatDir() error {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return err
	}
	return fileHandler.CreateDirectory(initiatPath)
}

func CreateEnvironmentDir(slug string) error {
	envPath, err := GetEnvironmentPath(slug)
	if err != nil {
		return err
	}
	return fileHandler.CreateDirectory(envPath)
}

func SetActiveEnvironment(slug string) error {
	activePath, err := GetActivePath()
	if err != nil {
		return err
	}

	if runtime.GOOS == WindowsOS {
		return fileHandler.WriteFile(activePath, slug)
	}

	envPath, err := GetEnvironmentPath(slug)
	if err != nil {
		return err
	}

	err = fileHandler.CreateSymlink(envPath, activePath)
	if err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	return nil
}

func UnsetActiveEnvironment() error {
	activePath, err := GetActivePath()
	if err != nil {
		return err
	}

	if runtime.GOOS == WindowsOS {
		// On Windows, write empty string to clear the active environment
		return fileHandler.WriteFile(activePath, "")
	}

	// On Unix-like systems, remove the symlink
	err = fileHandler.RemoveFile(activePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove active environment symlink: %w", err)
	}
	return nil
}

func GetActiveEnvironment() (string, error) {
	activePath, err := GetActivePath()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == WindowsOS {
		content, err := fileHandler.ReadFile(activePath)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(content), nil
	}

	target, err := fileHandler.ReadSymlink(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no active environment set")
		}
		return "", err
	}

	parts := strings.Split(target, string(filepath.Separator))
	return parts[len(parts)-1], nil
}

func ListLocalEnvironments() ([]EnvironmentInfo, error) {
	envsPath, err := GetEnvironmentsPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(envsPath)
	if err != nil {
		return nil, err
	}

	activeEnv, _ := GetActiveEnvironment()
	var envs []EnvironmentInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		envSlug := entry.Name()
		envs = append(envs, EnvironmentInfo{
			Slug:     envSlug,
			Name:     envSlug,
			IsActive: envSlug == activeEnv,
		})
	}

	return envs, nil
}

func LocalEnvironmentExists(slug string) bool {
	envPath, err := GetEnvironmentPath(slug)
	if err != nil {
		return false
	}
	_, err = os.Stat(envPath)
	return err == nil
}

func IsInitCompleted() bool {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(initiatPath)
	if err != nil {
		return false
	}

	// Check if gitignore is properly configured
	gitStatus, err := GetGitignoreStatus()
	if err != nil {
		return false
	}

	return gitStatus == GitignoreConfigured
}
