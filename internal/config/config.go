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

// Config holds the application configuration
type Config struct {
	APIBaseURL string `mapstructure:"api_base_url"`
}

var globalConfig *Config

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		APIBaseURL: "https://api.initflow.com",
	}
}

// InitConfig initializes the configuration system
func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".initflow")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	defaults := DefaultConfig()
	viper.SetDefault("api_base_url", defaults.APIBaseURL)

	viper.SetEnvPrefix("INITFLOW")
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

	configDir := filepath.Join(home, ".initflow")
	if err := os.MkdirAll(configDir, configDirPermissions); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configFile)
}
