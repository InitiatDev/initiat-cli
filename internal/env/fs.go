package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	InitiatDir      = ".initiat"
	EnvironmentsDir = "environments"
	ActiveFile      = "active"
	SecretsFile     = "secrets.env"
	EnvrcFile       = ".envrc"
	WindowsOS       = "windows"

	// File permissions
	DirPerms  = 0755
	FilePerms = 0600
)

func GetInitiatPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, InitiatDir), nil
}

func GetEnvironmentsPath() (string, error) {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(initiatPath, EnvironmentsDir), nil
}

func GetActivePath() (string, error) {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(initiatPath, ActiveFile), nil
}

func GetEnvironmentPath(slug string) (string, error) {
	envsPath, err := GetEnvironmentsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(envsPath, slug), nil
}

func GetSecretsPath(envSlug string) (string, error) {
	envPath, err := GetEnvironmentPath(envSlug)
	if err != nil {
		return "", err
	}
	return filepath.Join(envPath, SecretsFile), nil
}

func CreateInitiatDir() error {
	initiatPath, err := GetInitiatPath()
	if err != nil {
		return err
	}
	return os.MkdirAll(initiatPath, DirPerms)
}

func CreateEnvironmentDir(slug string) error {
	envPath, err := GetEnvironmentPath(slug)
	if err != nil {
		return err
	}
	return os.MkdirAll(envPath, DirPerms)
}

func SetActiveEnvironment(slug string) error {
	activePath, err := GetActivePath()
	if err != nil {
		return err
	}

	if runtime.GOOS == WindowsOS {
		return os.WriteFile(activePath, []byte(slug), FilePerms)
	}

	envPath, err := GetEnvironmentPath(slug)
	if err != nil {
		return err
	}

	_ = os.Remove(activePath) // Ignore error if file doesn't exist
	err = os.Symlink(envPath, activePath)
	if err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	return nil
}

func GetActiveEnvironment() (string, error) {
	activePath, err := GetActivePath()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == WindowsOS {
		// Validate path to prevent directory traversal
		if !filepath.IsAbs(activePath) {
			absPath, err := filepath.Abs(activePath)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path: %w", err)
			}
			activePath = absPath
		}

		content, err := os.ReadFile(activePath) // #nosec G304 - path is validated above
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	}

	target, err := os.Readlink(activePath)
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
		secretsPath := filepath.Join(envsPath, envSlug, SecretsFile)

		var info os.FileInfo
		hasSecrets := false
		if stat, err := os.Stat(secretsPath); err == nil {
			info = stat
			hasSecrets = true
		}

		var synced time.Time
		if hasSecrets {
			synced = info.ModTime()
		}

		envs = append(envs, EnvironmentInfo{
			Slug:       envSlug,
			Name:       envSlug,
			IsActive:   envSlug == activeEnv,
			Synced:     synced,
			HasSecrets: hasSecrets,
		})
	}

	return envs, nil
}

func WriteSecrets(envSlug string, content string) error {
	secretsPath, err := GetSecretsPath(envSlug)
	if err != nil {
		return err
	}

	if err := CreateEnvironmentDir(envSlug); err != nil {
		return err
	}

	return os.WriteFile(secretsPath, []byte(content), FilePerms)
}

func ReadSecrets(envSlug string) (string, error) {
	secretsPath, err := GetSecretsPath(envSlug)
	if err != nil {
		return "", err
	}

	// Validate path to prevent directory traversal
	if !filepath.IsAbs(secretsPath) {
		absPath, err := filepath.Abs(secretsPath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		secretsPath = absPath
	}

	content, err := os.ReadFile(secretsPath) // #nosec G304 - path is validated above
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return string(content), nil
}

func LocalEnvironmentExists(slug string) bool {
	envPath, err := GetEnvironmentPath(slug)
	if err != nil {
		return false
	}
	_, err = os.Stat(envPath)
	return err == nil
}
