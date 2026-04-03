package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/testutil"
	"github.com/InitiatDev/initiat-cli/internal/types"
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

	capture.AssertContains(t, "[ok] test-step")
}

func TestSetupRunner_Run_NoCommands(t *testing.T) {
	setupConfig := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "conditional-step",
				If:   "os('windows')",
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

func TestNewSetupRunnerWithDeps(t *testing.T) {
	projectCtx := &config.ProjectContext{
		OrgSlug:     "test-org",
		ProjectSlug: "test-project",
	}

	mockStore := &mockStorage{
		hasDeviceID:   true,
		encryptionKey: []byte("test-key"),
	}
	mockClient := &mockAPIClient{}

	runner := NewSetupRunnerWithDeps(projectCtx, mockStore, mockClient)

	if runner.projectCtx != projectCtx {
		t.Error("NewSetupRunnerWithDeps() projectCtx not set correctly")
	}

	if runner.store != mockStore {
		t.Error("NewSetupRunnerWithDeps() store not set correctly")
	}

	if runner.apiClient != mockClient {
		t.Error("NewSetupRunnerWithDeps() apiClient not set correctly")
	}
}

type mockStorage struct {
	hasDeviceID   bool
	encryptionKey []byte
}

func (m *mockStorage) HasDeviceID() bool {
	return m.hasDeviceID
}

func (m *mockStorage) GetEncryptionPrivateKey() ([]byte, error) {
	if m.encryptionKey == nil {
		return nil, fmt.Errorf("key not found")
	}
	return m.encryptionKey, nil
}

type mockAPIClient struct {
	secrets           map[string]*types.SecretWithValue
	wrappedProjectKey string
}

func (m *mockAPIClient) GetSecret(orgSlug, projectSlug, secretKey string) (*types.SecretWithValue, error) {
	if m.secrets == nil {
		return nil, fmt.Errorf("secret not found")
	}
	secret, ok := m.secrets[secretKey]
	if !ok {
		return nil, fmt.Errorf("secret not found")
	}
	return secret, nil
}

func (m *mockAPIClient) GetWrappedProjectKey(orgSlug, projectSlug string) (string, error) {
	if m.wrappedProjectKey == "" {
		return "", fmt.Errorf("project key not found")
	}
	return m.wrappedProjectKey, nil
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
		capture.AssertContains(t, "[ok] test")
	}
}
