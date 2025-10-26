package env

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckDirenvInstalled(t *testing.T) {
	installed := CheckDirenvInstalled()

	if !installed {
		t.Log("direnv not installed - this is expected in CI")
	}
}

func TestGetDirenvVersion(t *testing.T) {
	version, err := GetDirenvVersion()
	if err != nil {
		if !CheckDirenvInstalled() {
			t.Log("direnv not installed - skipping version test")
			return
		}
		t.Fatalf("GetDirenvVersion failed: %v", err)
	}

	if version == "" {
		t.Error("Expected non-empty version")
	}
}

func TestGenerateEnvrc(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := GenerateEnvrc()
	if err != nil {
		t.Fatalf("GenerateEnvrc failed: %v", err)
	}

	content, err := os.ReadFile(EnvrcFile)
	if err != nil {
		t.Fatalf("Failed to read .envrc: %v", err)
	}

	expectedContent := `dotenv ".initiat/active/secrets.env"
export INITIAT_ENV=$(basename "$(readlink .initiat/active 2>/dev/null || cat .initiat/active)")`

	if runtime.GOOS == "windows" {
		expectedContent = `dotenv ".initiat/active/secrets.env"
export INITIAT_ENV=$(cat .initiat/active)`
	}

	if string(content) != expectedContent {
		t.Errorf("Expected .envrc content:\n%s\nGot:\n%s", expectedContent, string(content))
	}
}

func TestGetInstallInstructions(t *testing.T) {
	instructions := GetInstallInstructions()

	if instructions == "" {
		t.Error("Expected non-empty install instructions")
	}

	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(instructions, "brew") {
			t.Error("Expected brew instructions for macOS")
		}
	case "linux":
		if !strings.Contains(instructions, "curl") {
			t.Error("Expected curl instructions for Linux")
		}
	case "windows":
		if !strings.Contains(instructions, "choco") {
			t.Error("Expected choco instructions for Windows")
		}
	}
}

func TestCheckDirenvHook(t *testing.T) {
	hasHook := CheckDirenvHook()

	if !hasHook {
		t.Log("direnv hook not found - this is expected in CI")
	}
}

func TestGenerateEnvrcToPath(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	customPath := "custom.envrc"
	err := GenerateEnvrcToPath(customPath)
	if err != nil {
		t.Fatalf("GenerateEnvrcToPath failed: %v", err)
	}

	content, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("Failed to read custom .envrc: %v", err)
	}

	expectedContent := `dotenv ".initiat/active/secrets.env"
export INITIAT_ENV=$(basename "$(readlink .initiat/active 2>/dev/null || cat .initiat/active)")`

	if runtime.GOOS == "windows" {
		expectedContent = `dotenv ".initiat/active/secrets.env"
export INITIAT_ENV=$(cat .initiat/active)`
	}

	if string(content) != expectedContent {
		t.Errorf("Expected .envrc content:\n%s\nGot:\n%s", expectedContent, string(content))
	}
}

func TestWriteDirenvHookToShellConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create a mock home directory
	homeDir := filepath.Join(tempDir, "home")
	err := os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}

	// Mock the home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	// Test zsh
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	err = WriteDirenvHookToShellConfig("zsh")
	if err != nil {
		t.Fatalf("WriteDirenvHookToShellConfig failed for zsh: %v", err)
	}

	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("Failed to read .zshrc: %v", err)
	}

	expectedHook := `eval "$(direnv hook zsh)"`
	if !strings.Contains(string(content), expectedHook) {
		t.Errorf("Expected direnv hook in .zshrc: %s", expectedHook)
	}

	// Test bash
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	err = WriteDirenvHookToShellConfig("bash")
	if err != nil {
		t.Fatalf("WriteDirenvHookToShellConfig failed for bash: %v", err)
	}

	content, err = os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("Failed to read .bashrc: %v", err)
	}

	expectedHook = `eval "$(direnv hook bash)"`
	if !strings.Contains(string(content), expectedHook) {
		t.Errorf("Expected direnv hook in .bashrc: %s", expectedHook)
	}
}

func TestWriteDirenvHookToShellConfigIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create a mock home directory
	homeDir := filepath.Join(tempDir, "home")
	err := os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}

	// Mock the home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	zshrcPath := filepath.Join(homeDir, ".zshrc")

	// Write hook first time
	err = WriteDirenvHookToShellConfig("zsh")
	if err != nil {
		t.Fatalf("First WriteDirenvHookToShellConfig failed: %v", err)
	}

	firstContent, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("Failed to read .zshrc after first write: %v", err)
	}

	// Write hook second time (should be idempotent)
	err = WriteDirenvHookToShellConfig("zsh")
	if err != nil {
		t.Fatalf("Second WriteDirenvHookToShellConfig failed: %v", err)
	}

	secondContent, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("Failed to read .zshrc after second write: %v", err)
	}

	if string(firstContent) != string(secondContent) {
		t.Error("WriteDirenvHookToShellConfig should be idempotent")
	}
}
