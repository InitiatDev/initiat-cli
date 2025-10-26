package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func CheckDirenvInstalled() bool {
	_, err := exec.LookPath("direnv")
	return err == nil
}

func GetDirenvVersion() (string, error) {
	if !CheckDirenvInstalled() {
		return "", fmt.Errorf("direnv not installed")
	}

	cmd := exec.Command("direnv", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get direnv version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func GenerateEnvrc() error {
	envrcPath := EnvrcFile

	content := `dotenv ".initiat/active/secrets.env"
export INITIAT_ENV=$(basename "$(readlink .initiat/active 2>/dev/null || cat .initiat/active)")`

	if runtime.GOOS == "windows" {
		content = `dotenv ".initiat/active/secrets.env"
export INITIAT_ENV=$(cat .initiat/active)`
	}

	return os.WriteFile(envrcPath, []byte(content), 0644)
}

func ReloadDirenv() error {
	if !CheckDirenvInstalled() {
		return fmt.Errorf("direnv not installed")
	}

	cmd := exec.Command("direnv", "reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AllowDirenv() error {
	if !CheckDirenvInstalled() {
		return fmt.Errorf("direnv not installed")
	}

	cmd := exec.Command("direnv", "allow")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install direnv"
	case "linux":
		return "curl -sfL https://direnv.net/install.sh | bash"
	case "windows":
		return "choco install direnv"
	default:
		return "Visit https://direnv.net/docs/installation.html"
	}
}

func CheckDirenvHook() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		return false
	}

	var configFile string
	if strings.Contains(shell, "zsh") {
		configFile = filepath.Join(homeDir, ".zshrc")
	} else if strings.Contains(shell, "bash") {
		configFile = filepath.Join(homeDir, ".bashrc")
	} else {
		return false
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "direnv")
}
