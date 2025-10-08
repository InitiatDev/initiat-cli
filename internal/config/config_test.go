package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "https://www.initiat.dev", cfg.API.BaseURL)
	assert.Equal(t, "30s", cfg.API.Timeout)
	assert.Equal(t, "initiat-cli", cfg.ServiceName)
	assert.Equal(t, "", cfg.Project.DefaultOrg)
	assert.Equal(t, "", cfg.Project.DefaultProject)
	assert.NotNil(t, cfg.Aliases)
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
	assert.Equal(t, "https://www.initiat.dev", cfg.API.BaseURL)
}

func TestInitConfig_WithConfigFile(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".initiat")
	err := os.MkdirAll(configDir, 0750)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `api:
  base_url: "http://localhost:4000"`
	err = os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err = InitConfig()
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:4000", cfg.API.BaseURL)
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
	assert.Equal(t, "http://localhost:3000", cfg.API.BaseURL)
}

func TestSet(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = Set("api.base_url", "http://localhost:8080")
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:8080", cfg.API.BaseURL)
}

func TestSave(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = Set("api.base_url", "http://localhost:9000")
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
	assert.Equal(t, "https://www.initiat.dev", cfg.API.BaseURL)
}

func TestAliasManagement(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = SetAlias("prod", "acme-corp/production")
	require.NoError(t, err)

	alias := GetAlias("prod")
	assert.Equal(t, "acme-corp/production", alias)

	nonExistent := GetAlias("nonexistent")
	assert.Equal(t, "", nonExistent)

	aliases := ListAliases()
	assert.Equal(t, "acme-corp/production", aliases["prod"])

	err = RemoveAlias("prod")
	require.NoError(t, err)

	alias = GetAlias("prod")
	assert.Equal(t, "", alias)
}

func TestProjectContextResolution(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = SetDefaultOrgSlug("default-org")
	require.NoError(t, err)
	err = SetDefaultProjectSlug("default-project")
	require.NoError(t, err)
	err = SetAlias("prod", "acme-corp/production")
	require.NoError(t, err)

	tests := []struct {
		name        string
		projectPath string
		org         string
		project     string
		expectedOrg string
		expectedWS  string
		expectError bool
	}{
		{
			name:        "explicit project path",
			projectPath: "acme-corp/production",
			expectedOrg: "acme-corp",
			expectedWS:  "production",
		},
		{
			name:        "project path via alias",
			projectPath: "prod",
			expectedOrg: "acme-corp",
			expectedWS:  "production",
		},
		{
			name:        "explicit org and project",
			org:         "test-org",
			project:     "test-project",
			expectedOrg: "test-org",
			expectedWS:  "test-project",
		},
		{
			name:        "default org with explicit project",
			project:     "staging",
			expectedOrg: "default-org",
			expectedWS:  "staging",
		},
		{
			name:        "full defaults",
			expectedOrg: "default-org",
			expectedWS:  "default-project",
		},
		{
			name:        "invalid project path format",
			projectPath: "invalid-format",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := ResolveProjectContext(tt.projectPath, tt.org, tt.project)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ctx)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ctx)
				assert.Equal(t, tt.expectedOrg, ctx.OrgSlug)
				assert.Equal(t, tt.expectedWS, ctx.ProjectSlug)
				assert.Equal(t, fmt.Sprintf("%s/%s", tt.expectedOrg, tt.expectedWS), ctx.String())
			}
		})
	}
}

func TestProjectContextResolution_ErrorCases(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	ctx, err := ResolveProjectContext("", "", "test-project")
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.Contains(t, err.Error(), "no default organization configured")

	ctx, err = ResolveProjectContext("", "", "")
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.Contains(t, err.Error(), "no project context available")
}

func TestResetToDefaults(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := InitConfig()
	require.NoError(t, err)

	err = Set("api.base_url", "http://localhost:8080")
	require.NoError(t, err)
	err = Set("api.timeout", "60s")
	require.NoError(t, err)
	err = Set("project.default_org", "test-org")
	require.NoError(t, err)
	err = Set("project.default_project", "test-project")
	require.NoError(t, err)
	err = SetAlias("prod", "acme-corp/production")
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "http://localhost:8080", cfg.API.BaseURL)
	assert.Equal(t, "60s", cfg.API.Timeout)
	assert.Equal(t, "test-org", cfg.Project.DefaultOrg)
	assert.Equal(t, "test-project", cfg.Project.DefaultProject)
	assert.Equal(t, "acme-corp/production", GetAlias("prod"))

	err = ResetToDefaults()
	require.NoError(t, err)

	cfg = Get()
	assert.Equal(t, "https://www.initiat.dev", cfg.API.BaseURL)
	assert.Equal(t, "30s", cfg.API.Timeout)
	assert.Equal(t, "initiat-cli", cfg.ServiceName)
	assert.Equal(t, "", cfg.Project.DefaultOrg)
	assert.Equal(t, "", cfg.Project.DefaultProject)
	assert.Equal(t, "", GetAlias("prod"))
	assert.Empty(t, ListAliases())
}

func TestFindLocalConfig_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	localConfig, err := FindLocalConfig()
	require.NoError(t, err)
	assert.Nil(t, localConfig)
}

func TestFindLocalConfig_WithFile(t *testing.T) {
	tmpDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	initiatContent := `org: test-org
project: test-project`
	err = os.WriteFile(".initiat", []byte(initiatContent), 0600)
	require.NoError(t, err)

	localConfig, err := FindLocalConfig()
	require.NoError(t, err)
	require.NotNil(t, localConfig)
	assert.Equal(t, "test-org", localConfig.Org)
	assert.Equal(t, "test-project", localConfig.Project)
}

func TestFindLocalConfig_WithComments(t *testing.T) {
	tmpDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	initiatContent := `# This is a comment
org: test-org

# Another comment
project: test-project
# End comment`
	err = os.WriteFile(".initiat", []byte(initiatContent), 0600)
	require.NoError(t, err)

	localConfig, err := FindLocalConfig()
	require.NoError(t, err)
	require.NotNil(t, localConfig)
	assert.Equal(t, "test-org", localConfig.Org)
	assert.Equal(t, "test-project", localConfig.Project)
}

func TestResolveProjectContext_WithLocalConfig(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	testDir := filepath.Join(tmpDir, "test-with-local-config")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(testDir)
	require.NoError(t, err)

	err = InitConfig()
	require.NoError(t, err)

	initiatContent := `org: local-org
project: local-project`
	err = os.WriteFile(".initiat", []byte(initiatContent), 0600)
	require.NoError(t, err)

	ctx, err := ResolveProjectContext("", "", "")
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.Equal(t, "local-org", ctx.OrgSlug)
	assert.Equal(t, "local-project", ctx.ProjectSlug)
}

func TestResolveProjectContext_LocalConfigPriority(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	testDir := filepath.Join(tmpDir, "test-local-config")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(testDir)
	require.NoError(t, err)

	err = InitConfig()
	require.NoError(t, err)

	err = SetDefaultOrgSlug("global-org")
	require.NoError(t, err)
	err = SetDefaultProjectSlug("global-project")
	require.NoError(t, err)

	initiatContent := `org: local-org
project: local-project`
	err = os.WriteFile(".initiat", []byte(initiatContent), 0600)
	require.NoError(t, err)

	ctx, err := ResolveProjectContext("", "", "")
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.Equal(t, "local-org", ctx.OrgSlug)
	assert.Equal(t, "local-project", ctx.ProjectSlug)
}

func TestResolveProjectContext_FlagsOverrideLocalConfig(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	testDir := filepath.Join(tmpDir, "test-flags-override")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(testDir)
	require.NoError(t, err)

	err = InitConfig()
	require.NoError(t, err)

	initiatContent := `org: local-org
project: local-project`
	err = os.WriteFile(".initiat", []byte(initiatContent), 0600)
	require.NoError(t, err)

	ctx, err := ResolveProjectContext("", "flag-org", "flag-project")
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.Equal(t, "flag-org", ctx.OrgSlug)
	assert.Equal(t, "flag-project", ctx.ProjectSlug)
}

func TestResolveProjectContext_IncompleteLocalConfig(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	testDir := filepath.Join(tmpDir, "test-incomplete-config")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(testDir)
	require.NoError(t, err)

	err = InitConfig()
	require.NoError(t, err)

	initiatContent := `org: local-org`
	err = os.WriteFile(".initiat", []byte(initiatContent), 0600)
	require.NoError(t, err)

	ctx, err := ResolveProjectContext("", "", "")
	require.Error(t, err)
	assert.Nil(t, ctx)
	assert.Contains(t, err.Error(), "no project context available")
}
