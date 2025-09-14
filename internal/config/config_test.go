package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "https://api.initflow.com", cfg.APIBaseURL)
}

func TestInitConfig_WithDefaults(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create temporary directory for config
	tmpDir := t.TempDir()

	// Set home directory to temp dir
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "https://api.initflow.com", cfg.APIBaseURL)
}

func TestInitConfig_WithConfigFile(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create temporary directory for config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".initflow")
	err := os.MkdirAll(configDir, 0750)
	require.NoError(t, err)

	// Create config file
	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `api_base_url: "http://localhost:4000"`
	err = os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set home directory to temp dir
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err = InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:4000", cfg.APIBaseURL)
}

func TestInitConfig_WithEnvironmentVariable(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create temporary directory for config
	tmpDir := t.TempDir()

	// Set environment variable
	originalEnv := os.Getenv("INITFLOW_API_BASE_URL")
	_ = os.Setenv("INITFLOW_API_BASE_URL", "http://localhost:3000")
	defer func() {
		if originalEnv == "" {
			_ = os.Unsetenv("INITFLOW_API_BASE_URL")
		} else {
			_ = os.Setenv("INITFLOW_API_BASE_URL", originalEnv)
		}
	}()

	// Set home directory to temp dir
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:3000", cfg.APIBaseURL)
}

func TestSet(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create temporary directory for config
	tmpDir := t.TempDir()

	// Set home directory to temp dir
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	// Test setting a value
	err = Set("api_base_url", "http://localhost:8080")
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:8080", cfg.APIBaseURL)
}

func TestSave(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create temporary directory for config
	tmpDir := t.TempDir()

	// Set home directory to temp dir
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	// Set a custom value
	err = Set("api_base_url", "http://localhost:9000")
	require.NoError(t, err)

	// Save config
	err = Save()
	require.NoError(t, err)

	// Verify config file was created
	configFile := filepath.Join(tmpDir, ".initflow", "config.yaml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(configFile) // #nosec G304 - test file path is controlled
	require.NoError(t, err)
	assert.Contains(t, string(content), "http://localhost:9000")
}

func TestGet_BeforeInit(t *testing.T) {
	// Reset global config
	globalConfig = nil

	cfg := Get()
	assert.Equal(t, "https://api.initflow.com", cfg.APIBaseURL)
}
