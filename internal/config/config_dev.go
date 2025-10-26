//go:build dev

package config

const defaultAPIBaseURL = "http://localhost:4000"

// GetAPIBaseURL returns localhost:4000 for dev builds, ignoring any config file settings
func GetAPIBaseURL() string {
	return "http://localhost:4000"
}
