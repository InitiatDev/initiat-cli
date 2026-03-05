package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/testutil"
)

func TestRunSetupRun_AgentEnabled_UserDeclinesAgentMode(t *testing.T) {
	if err := config.InitConfig(); err != nil {
		t.Fatalf("init config: %v", err)
	}
	if err := config.Set("agent.enabled", true); err != nil {
		t.Fatalf("enable agent: %v", err)
	}

	stdin, err := testutil.MockStdinWithLines("n")
	if err != nil {
		t.Fatalf("mock stdin: %v", err)
	}
	defer stdin.Restore()

	tmpDir := t.TempDir()
	setupFile := filepath.Join(tmpDir, "setup.yml")

	yamlContent := `version: 1
name: "Test Setup"
setup:
  - name: "fail"
    run: "sh -c 'exit 1'"
`

	if err := os.WriteFile(setupFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write setup file: %v", err)
	}

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	err = runSetupRun(setupRunCmd, []string{setupFile})
	if err == nil {
		t.Fatalf("expected error")
	}

	capture.AssertContains(t, "Enter agent mode?")
}
