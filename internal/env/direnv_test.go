package env

import (
	"os"
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
