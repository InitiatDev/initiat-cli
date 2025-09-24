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
	assert.Equal(t, "https://www.initiat.dev", cfg.APIBaseURL)
}

func TestInitConfig_WithDefaults(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "https://www.initiat.dev", cfg.APIBaseURL)
}

func TestInitConfig_WithConfigFile(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".initiat")
	err := os.MkdirAll(configDir, 0750)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `api_base_url: "http://localhost:4000"`
	err = os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err = InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:4000", cfg.APIBaseURL)
}

func TestInitConfig_WithEnvironmentVariable(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalEnv := os.Getenv("INITIAT_API_BASE_URL")
	_ = os.Setenv("INITIAT_API_BASE_URL", "http://localhost:3000")
	defer func() {
		if originalEnv == "" {
			_ = os.Unsetenv("INITIAT_API_BASE_URL")
		} else {
			_ = os.Setenv("INITIAT_API_BASE_URL", originalEnv)
		}
	}()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:3000", cfg.APIBaseURL)
}

func TestSet(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = Set("api_base_url", "http://localhost:8080")
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:8080", cfg.APIBaseURL)
}

func TestSave(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = Set("api_base_url", "http://localhost:9000")
	require.NoError(t, err)

	err = Save()
	require.NoError(t, err)

	configFile := filepath.Join(tmpDir, ".initiat", "config.yaml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	content, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "http://localhost:9000")
}

func TestGet_BeforeInit(t *testing.T) {
	globalConfig = nil

	cfg := Get()
	assert.Equal(t, "https://www.initiat.dev", cfg.APIBaseURL)
}
