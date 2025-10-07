package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	configDirPermissions     = 0750
	projectPathPartsExpected = 2
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
	{"org", "project.default_org", ""},
	{"project", "project.default_project", ""},
}

type Config struct {
	API         APIConfig         `mapstructure:"api"`
	Project     ProjectConfig     `mapstructure:"project"`
	Aliases     map[string]string `mapstructure:"aliases"`
	ServiceName string            `mapstructure:"service_name"`
}

type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout string `mapstructure:"timeout"`
}

type ProjectConfig struct {
	DefaultOrg     string `mapstructure:"default_org"`
	DefaultProject string `mapstructure:"default_project"`
}

var globalConfig *Config

func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			BaseURL: defaultAPIBaseURL,
			Timeout: "30s",
		},
		Project: ProjectConfig{
			DefaultOrg:     "",
			DefaultProject: "",
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
	viper.SetDefault("project.default_org", defaults.Project.DefaultOrg)
	viper.SetDefault("project.default_project", defaults.Project.DefaultProject)
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
	return cfg.Project.DefaultOrg
}

func GetDefaultProjectSlug() string {
	cfg := Get()
	return cfg.Project.DefaultProject
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

func SetAlias(alias, projectPath string) error {
	cfg := Get()
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	if err := Set("aliases."+alias, projectPath); err != nil {
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
	if err := Set("project.default_org", orgSlug); err != nil {
		return fmt.Errorf("failed to set default org slug: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func SetDefaultProjectSlug(projectSlug string) error {
	if err := Set("project.default_project", projectSlug); err != nil {
		return fmt.Errorf("failed to set default project slug: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func ClearDefaultOrgSlug() error {
	return SetDefaultOrgSlug("")
}

func ClearDefaultProjectSlug() error {
	return SetDefaultProjectSlug("")
}

type ProjectContext struct {
	OrgSlug     string
	ProjectSlug string
}

func (w ProjectContext) String() string {
	return fmt.Sprintf("%s/%s", w.OrgSlug, w.ProjectSlug)
}

// ResolveProjectContext resolves project context based on priority:
// 1. projectPath (full path or alias) 2. org + project 3. default org + project 4. full defaults
func ResolveProjectContext(projectPath, org, project string) (*ProjectContext, error) {
	if projectPath != "" {
		if aliasPath := GetAlias(projectPath); aliasPath != "" {
			projectPath = aliasPath
		}

		parts := strings.Split(projectPath, "/")
		if len(parts) != projectPathPartsExpected {
			return nil, fmt.Errorf("project path must be in format 'org/project', got: %s", projectPath)
		}

		return &ProjectContext{
			OrgSlug:     parts[0],
			ProjectSlug: parts[1],
		}, nil
	}

	if org != "" && project != "" {
		return &ProjectContext{
			OrgSlug:     org,
			ProjectSlug: project,
		}, nil
	}

	if project != "" {
		defaultOrg := GetDefaultOrgSlug()
		if defaultOrg == "" {
			return nil, fmt.Errorf("project specified but no default organization configured. " +
				"Use 'initiat config set org <org>' or specify --org")
		}

		return &ProjectContext{
			OrgSlug:     defaultOrg,
			ProjectSlug: project,
		}, nil
	}

	defaultOrg := GetDefaultOrgSlug()
	defaultProject := GetDefaultProjectSlug()

	if defaultOrg != "" && defaultProject != "" {
		return &ProjectContext{
			OrgSlug:     defaultOrg,
			ProjectSlug: defaultProject,
		}, nil
	}

	switch {
	case defaultOrg == "" && defaultProject == "":
		return nil, fmt.Errorf("no project context available. Specify --project-path, --org and --project, " +
			"or configure defaults with 'initiat config set org <org>' and 'initiat config set project <project>'")
	case defaultOrg == "":
		return nil, fmt.Errorf("no organization context available. Specify --project-path, --org, " +
			"or configure default with 'initiat config set org <org>'")
	default:
		return nil, fmt.Errorf("no project context available. Specify --project-path, --project, " +
			"or configure default with 'initiat config set project <project>'")
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
