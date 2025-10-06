package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	configDirPermissions       = 0750
	workspacePathPartsExpected = 2
)

var defaultAPIBaseURL = "https://www.initiat.dev"

type ConfigKey struct {
	Simplified string
	Actual     string
	Default    interface{}
}

var ConfigKeys = []ConfigKey{
	{"api.url", "api.base_url", defaultAPIBaseURL},
	{"api.timeout", "api.timeout", "30s"},
	{"service", "service_name", "initiat-cli"},
	{"org", "workspace.default_org", ""},
	{"workspace", "workspace.default_workspace", ""},
}

type Config struct {
	API         APIConfig         `mapstructure:"api"`
	Workspace   WorkspaceConfig   `mapstructure:"workspace"`
	Aliases     map[string]string `mapstructure:"aliases"`
	ServiceName string            `mapstructure:"service_name"`
}

type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout string `mapstructure:"timeout"`
}

type WorkspaceConfig struct {
	DefaultOrg       string `mapstructure:"default_org"`
	DefaultWorkspace string `mapstructure:"default_workspace"`
}

var globalConfig *Config

func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			BaseURL: defaultAPIBaseURL,
			Timeout: "30s",
		},
		Workspace: WorkspaceConfig{
			DefaultOrg:       "",
			DefaultWorkspace: "",
		},
		Aliases:     make(map[string]string),
		ServiceName: "initiat-cli",
	}
}

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".initiat")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	defaults := DefaultConfig()
	viper.SetDefault("api.base_url", defaultAPIBaseURL)
	viper.SetDefault("api.timeout", defaults.API.Timeout)
	viper.SetDefault("service_name", defaults.ServiceName)
	viper.SetDefault("workspace.default_org", defaults.Workspace.DefaultOrg)
	viper.SetDefault("workspace.default_workspace", defaults.Workspace.DefaultWorkspace)
	viper.SetDefault("aliases", defaults.Aliases)

	viper.SetEnvPrefix("INITIAT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func Get() *Config {
	if globalConfig == nil {
		return DefaultConfig()
	}
	return globalConfig
}

func Set(key string, value interface{}) error {
	viper.Set(key, value)

	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

func Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".initiat")
	if err := os.MkdirAll(configDir, configDirPermissions); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configFile)
}

func GetDefaultOrgSlug() string {
	cfg := Get()
	return cfg.Workspace.DefaultOrg
}

func GetDefaultWorkspaceSlug() string {
	cfg := Get()
	return cfg.Workspace.DefaultWorkspace
}

func GetAPIBaseURL() string {
	cfg := Get()
	return cfg.API.BaseURL
}

func GetAPITimeout() string {
	cfg := Get()
	return cfg.API.Timeout
}

func GetServiceName() string {
	cfg := Get()
	return cfg.ServiceName
}

func GetAlias(alias string) string {
	aliases := viper.GetStringMapString("aliases")
	return aliases[alias]
}

func SetAlias(alias, workspacePath string) error {
	cfg := Get()
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	if err := Set("aliases."+alias, workspacePath); err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func RemoveAlias(alias string) error {
	cfg := Get()
	if cfg.Aliases == nil {
		return nil
	}

	if _, exists := cfg.Aliases[alias]; !exists {
		return nil
	}

	aliases := make(map[string]string)
	for k, v := range cfg.Aliases {
		if k != alias {
			aliases[k] = v
		}
	}

	viper.Set("aliases", aliases)

	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func ListAliases() map[string]string {
	return viper.GetStringMapString("aliases")
}

func SetDefaultOrgSlug(orgSlug string) error {
	if err := Set("workspace.default_org", orgSlug); err != nil {
		return fmt.Errorf("failed to set default org slug: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func SetDefaultWorkspaceSlug(workspaceSlug string) error {
	if err := Set("workspace.default_workspace", workspaceSlug); err != nil {
		return fmt.Errorf("failed to set default workspace slug: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func ClearDefaultOrgSlug() error {
	return SetDefaultOrgSlug("")
}

func ClearDefaultWorkspaceSlug() error {
	return SetDefaultWorkspaceSlug("")
}

type WorkspaceContext struct {
	OrgSlug       string
	WorkspaceSlug string
}

func (w WorkspaceContext) String() string {
	return fmt.Sprintf("%s/%s", w.OrgSlug, w.WorkspaceSlug)
}

// ResolveWorkspaceContext resolves workspace context based on priority:
// 1. workspacePath (full path or alias) 2. org + workspace 3. default org + workspace 4. full defaults
func ResolveWorkspaceContext(workspacePath, org, workspace string) (*WorkspaceContext, error) {
	if workspacePath != "" {
		if aliasPath := GetAlias(workspacePath); aliasPath != "" {
			workspacePath = aliasPath
		}

		parts := strings.Split(workspacePath, "/")
		if len(parts) != workspacePathPartsExpected {
			return nil, fmt.Errorf("workspace path must be in format 'org/workspace', got: %s", workspacePath)
		}

		return &WorkspaceContext{
			OrgSlug:       parts[0],
			WorkspaceSlug: parts[1],
		}, nil
	}

	if org != "" && workspace != "" {
		return &WorkspaceContext{
			OrgSlug:       org,
			WorkspaceSlug: workspace,
		}, nil
	}

	if workspace != "" {
		defaultOrg := GetDefaultOrgSlug()
		if defaultOrg == "" {
			return nil, fmt.Errorf("workspace specified but no default organization configured. " +
				"Use 'initiat config set org <org>' or specify --org")
		}

		return &WorkspaceContext{
			OrgSlug:       defaultOrg,
			WorkspaceSlug: workspace,
		}, nil
	}

	defaultOrg := GetDefaultOrgSlug()
	defaultWorkspace := GetDefaultWorkspaceSlug()

	if defaultOrg != "" && defaultWorkspace != "" {
		return &WorkspaceContext{
			OrgSlug:       defaultOrg,
			WorkspaceSlug: defaultWorkspace,
		}, nil
	}

	switch {
	case defaultOrg == "" && defaultWorkspace == "":
		return nil, fmt.Errorf("no workspace context available. Specify --workspace-path, --org and --workspace, " +
			"or configure defaults with 'initiat config set org <org>' and 'initiat config set workspace <workspace>'")
	case defaultOrg == "":
		return nil, fmt.Errorf("no organization context available. Specify --workspace-path, --org, " +
			"or configure default with 'initiat config set org <org>'")
	default:
		return nil, fmt.Errorf("no workspace context available. Specify --workspace-path, --workspace, " +
			"or configure default with 'initiat config set workspace <workspace>'")
	}
}

func EnsureConfigFileExists() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configFile := filepath.Join(home, ".initiat", "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("📁 Config file not found at %s\n", configFile)
		fmt.Print("❓ Create configuration file? (y/N): ")

		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			return fmt.Errorf("configuration file creation cancelled")
		}

		fmt.Printf("✅ Creating config file at %s\n", configFile)
		return Save()
	}
	return nil
}

func FindConfigKey(simplified string) *ConfigKey {
	for _, key := range ConfigKeys {
		if key.Simplified == simplified {
			return &key
		}
	}
	return nil
}

func MapSimplifiedKey(simplified string) string {
	if key := FindConfigKey(simplified); key != nil {
		return key.Actual
	}
	return simplified
}

func IsValidConfigKey(simplified string) bool {
	return FindConfigKey(simplified) != nil
}

func GetAllConfigKeys() []string {
	keys := make([]string, len(ConfigKeys))
	for i, key := range ConfigKeys {
		keys[i] = key.Simplified
	}
	return keys
}

func ResetToDefaults() error {
	for _, key := range ConfigKeys {
		if err := Set(key.Actual, key.Default); err != nil {
			return fmt.Errorf("failed to reset %s: %w", key.Simplified, err)
		}
	}

	if err := Set("aliases", make(map[string]string)); err != nil {
		return fmt.Errorf("failed to reset aliases: %w", err)
	}

	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save reset configuration: %w", err)
	}

	return nil
}
