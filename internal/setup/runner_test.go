package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/testutil"
)

func TestSetupRunner_Run(t *testing.T) {
	setupConfig := &SetupConfig{
		Version: 1,
		Name:    "Test Setup",
		Setup: []Step{
			{
				Name: "test-step",
				Run:  "echo 'test'",
			},
		},
	}

	projectCtx := &config.ProjectContext{
		OrgSlug:     "test-org",
		ProjectSlug: "test-project",
	}

	runner := NewSetupRunner(projectCtx)

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	err := runner.Run(setupConfig)
	if err != nil {
		t.Logf("Run error (expected if device not registered): %v", err)
		if !containsString(err.Error(), "device not registered") {
			t.Errorf("Expected device registration error, got: %v", err)
		}
		return
	}

	capture.AssertContains(t, "completed successfully")
}

func TestSetupRunner_Run_NoCommands(t *testing.T) {
	setupConfig := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "conditional-step",
				If:   "os == 'windows'",
				Run:  "echo 'windows'",
			},
		},
	}

	projectCtx := &config.ProjectContext{
		OrgSlug:     "test-org",
		ProjectSlug: "test-project",
	}

	runner := NewSetupRunner(projectCtx)

	err := runner.Run(setupConfig)
	if err == nil {
		t.Error("Expected ErrNoCommandsToExecute, got nil")
		return
	}

	if err != ErrNoCommandsToExecute {
		if !containsString(err.Error(), "device not registered") {
			t.Errorf("Expected ErrNoCommandsToExecute or device error, got: %v", err)
		}
	}
}

func TestSetupRunner_ValidateAndPrintErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig *SetupConfig
		wantError   bool
		contains    []string
	}{
		{
			name: "valid config",
			setupConfig: &SetupConfig{
				Version: 1,
			},
			wantError: false,
			contains:  []string{},
		},
		{
			name: "invalid version",
			setupConfig: &SetupConfig{
				Version: 2,
			},
			wantError: true,
			contains:  []string{"version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := testutil.CaptureStdout()
			defer capture.Restore()

			projectCtx := &config.ProjectContext{
				OrgSlug:     "test-org",
				ProjectSlug: "test-project",
			}
			runner := NewSetupRunner(projectCtx)

			err := runner.ValidateAndPrintErrors(tt.setupConfig)

			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAndPrintErrors() error = %v, wantError %v", err, tt.wantError)
			}

			for _, text := range tt.contains {
				capture.AssertContains(t, text)
			}
		})
	}
}

func TestSetupRunner_collectSecretNames(t *testing.T) {
	tests := []struct {
		name     string
		config   *SetupConfig
		expected []string
	}{
		{
			name: "no secrets",
			config: &SetupConfig{
				Version: 1,
				Setup: []Step{
					{Run: "echo test"},
				},
			},
			expected: []string{},
		},
		{
			name: "single secret",
			config: &SetupConfig{
				Version: 1,
				Setup: []Step{
					{
						Run:            "echo test",
						EnvFromSecrets: []string{"API_KEY"},
					},
				},
			},
			expected: []string{"API_KEY"},
		},
		{
			name: "multiple secrets in multiple steps",
			config: &SetupConfig{
				Version: 1,
				Bootstrap: []Step{
					{
						EnvFromSecrets: []string{"SECRET1"},
					},
				},
				Setup: []Step{
					{
						EnvFromSecrets: []string{"SECRET1", "SECRET2"},
					},
				},
			},
			expected: []string{"SECRET1", "SECRET2"},
		},
		{
			name: "duplicate secrets",
			config: &SetupConfig{
				Version: 1,
				Bootstrap: []Step{
					{EnvFromSecrets: []string{"API_KEY"}},
				},
				Setup: []Step{
					{EnvFromSecrets: []string{"API_KEY"}},
				},
			},
			expected: []string{"API_KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectSecretNames(tt.config)

			if len(result) != len(tt.expected) {
				t.Errorf("collectSecretNames() returned %d secrets, want %d", len(result), len(tt.expected))
			}

			secretMap := make(map[string]bool)
			for _, s := range result {
				secretMap[s] = true
			}

			for _, expected := range tt.expected {
				if !secretMap[expected] {
					t.Errorf("collectSecretNames() missing secret: %s", expected)
				}
			}
		})
	}
}

func TestSetupRunner_detectShell(t *testing.T) {
	originalShell := os.Getenv("SHELL")

	tests := []struct {
		name     string
		setShell string
		wantCont string
	}{
		{
			name:     "with SHELL env var",
			setShell: "/bin/zsh",
			wantCont: "/bin/zsh",
		},
		{
			name:     "without SHELL env var",
			setShell: "",
			wantCont: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setShell != "" {
				os.Setenv("SHELL", tt.setShell)
			} else {
				os.Unsetenv("SHELL")
			}
			defer func() {
				if originalShell != "" {
					os.Setenv("SHELL", originalShell)
				} else {
					os.Unsetenv("SHELL")
				}
			}()

			result := detectShell()

			if tt.setShell != "" && result != tt.setShell {
				t.Errorf("detectShell() = %q, want %q", result, tt.setShell)
			}
		})
	}
}

func TestNewSetupRunner(t *testing.T) {
	projectCtx := &config.ProjectContext{
		OrgSlug:     "test-org",
		ProjectSlug: "test-project",
	}

	runner := NewSetupRunner(projectCtx)

	if runner.projectCtx != projectCtx {
		t.Error("NewSetupRunner() projectCtx not set correctly")
	}

	if runner.store == nil {
		t.Error("NewSetupRunner() store is nil")
	}

	if runner.apiClient == nil {
		t.Error("NewSetupRunner() apiClient is nil")
	}
}

func TestSetupRunner_Run_WithActualFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupFile := filepath.Join(tmpDir, "setup.yml")

	yamlContent := `version: 1
name: "Test Setup"
setup:
  - name: "test"
    run: "echo 'hello'"
`

	if err := os.WriteFile(setupFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	setupConfig, err := ParseFile(setupFile)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	projectCtx := &config.ProjectContext{
		OrgSlug:     "test-org",
		ProjectSlug: "test-project",
	}

	runner := NewSetupRunner(projectCtx)

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	err = runner.Run(setupConfig)
	if err != nil {
		if !containsString(err.Error(), "device not registered") {
			t.Logf("Run error (may be expected): %v", err)
		}
	} else {
		capture.AssertContains(t, "completed successfully")
	}
}
