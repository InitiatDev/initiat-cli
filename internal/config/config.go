package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configDirPermissions = 0750
)

type Config struct {
	APIBaseURL     string `mapstructure:"api_base_url"`
	ServiceName    string `mapstructure:"service_name"`
	DefaultOrgSlug string `mapstructure:"default_org_slug"`
}

var globalConfig *Config

func DefaultConfig() *Config {
	return &Config{
		APIBaseURL:     "https://www.initiat.dev",
		ServiceName:    "initiat-cli",
		DefaultOrgSlug: "",
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
	viper.SetDefault("api_base_url", defaults.APIBaseURL)
	viper.SetDefault("service_name", defaults.ServiceName)
	viper.SetDefault("default_org_slug", defaults.DefaultOrgSlug)

	viper.SetEnvPrefix("INITIAT")
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

// GetDefaultOrgSlug returns the default organization slug from config
func GetDefaultOrgSlug() string {
	cfg := Get()
	return cfg.DefaultOrgSlug
}

// SetDefaultOrgSlug sets the default organization slug and saves the config
func SetDefaultOrgSlug(orgSlug string) error {
	if err := Set("default_org_slug", orgSlug); err != nil {
		return fmt.Errorf("failed to set default org slug: %w", err)
	}

	if err := Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ClearDefaultOrgSlug removes the default organization slug and saves the config
func ClearDefaultOrgSlug() error {
	return SetDefaultOrgSlug("")
}
