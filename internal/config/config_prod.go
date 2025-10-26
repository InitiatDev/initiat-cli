//go:build !dev

package config

const defaultAPIBaseURL = "https://www.initiat.dev"

// GetAPIBaseURL returns the configured API base URL for production builds
func GetAPIBaseURL() string {
	cfg := Get()
	return cfg.API.BaseURL
}
