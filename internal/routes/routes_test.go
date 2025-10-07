package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "v1", APIVersion)
	assert.Equal(t, "/api/v1", APIBasePath)
}

func TestAuthRoutes(t *testing.T) {
	assert.Equal(t, "/api/v1/auth/login", AuthLogin)
}

func TestDeviceRoutes(t *testing.T) {
	assert.Equal(t, "/api/v1/devices", Devices)
}

func TestProjectRoutes(t *testing.T) {
	assert.Equal(t, "/api/v1/projects", Projects)
}

func TestProjectRoutes_InitializeKey(t *testing.T) {
	route := Project.InitializeKey("acme-corp", "production")
	assert.Equal(t, "/api/v1/projects/acme-corp/production/initialize", route)
}

func TestProjectRoutes_GetBySlug(t *testing.T) {
	route := Project.GetBySlug("acme-corp", "staging")
	assert.Equal(t, "/api/v1/projects/acme-corp/staging", route)
}

func TestProjectRoutes_Secrets(t *testing.T) {
	route := Project.Secrets("my-org", "development")
	assert.Equal(t, "/api/v1/projects/my-org/development/secrets", route)
}

func TestProjectRoutes_SecretByKey(t *testing.T) {
	route := Project.SecretByKey("acme-corp", "production", "API_KEY")
	assert.Equal(t, "/api/v1/projects/acme-corp/production/secrets/API_KEY", route)
}

func TestProjectRoutes_InviteDevice(t *testing.T) {
	route := Project.InviteDevice("my-org", "staging")
	assert.Equal(t, "/api/v1/projects/my-org/staging/invite-device", route)
}

func TestDeviceRoutes_GetByID(t *testing.T) {
	route := Device.GetByID("abc123")
	assert.Equal(t, "/api/v1/devices/abc123", route)
}

func TestDeviceRoutes_Revoke(t *testing.T) {
	route := Device.Revoke("def456")
	assert.Equal(t, "/api/v1/devices/def456/revoke", route)
}

func TestHTTPMethods(t *testing.T) {
	assert.Equal(t, "GET", GET)
	assert.Equal(t, "POST", POST)
	assert.Equal(t, "PUT", PUT)
	assert.Equal(t, "DELETE", DELETE)
	assert.Equal(t, "PATCH", PATCH)
}

func TestPredefinedRoutes(t *testing.T) {
	assert.Equal(t, Route{Method: POST, Path: "/api/v1/auth/login"}, LoginRoute)
	assert.Equal(t, Route{Method: GET, Path: "/api/v1/projects"}, ListProjectsRoute)
	assert.Equal(t, Route{Method: POST, Path: "/api/v1/devices"}, RegisterDeviceRoute)
	assert.Equal(t, Route{Method: GET, Path: "/api/v1/devices"}, ListDevicesRoute)
}

func TestBuildURL(t *testing.T) {
	baseURL := "https://www.initiat.dev"

	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
	}{
		{
			name:     "login route",
			baseURL:  baseURL,
			path:     AuthLogin,
			expected: "https://www.initiat.dev/api/v1/auth/login",
		},
		{
			name:     "projects route",
			baseURL:  baseURL,
			path:     Projects,
			expected: "https://www.initiat.dev/api/v1/projects",
		},
		{
			name:     "localhost development",
			baseURL:  "http://localhost:4000",
			path:     AuthLogin,
			expected: "http://localhost:4000/api/v1/auth/login",
		},
		{
			name:     "project initialize key",
			baseURL:  baseURL,
			path:     Project.InitializeKey("acme-corp", "production"),
			expected: "https://www.initiat.dev/api/v1/projects/acme-corp/production/initialize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildURL(tt.baseURL, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRouteStruct(t *testing.T) {
	route := Route{
		Method: POST,
		Path:   AuthLogin,
	}

	assert.Equal(t, "POST", route.Method)
	assert.Equal(t, "/api/v1/auth/login", route.Path)
}

func TestProjectRoutesWithDifferentSlugs(t *testing.T) {
	testCases := []struct {
		orgSlug     string
		projectSlug string
		expected    string
	}{
		{"acme", "prod", "/api/v1/projects/acme/prod/initialize"},
		{"my-company", "staging-env", "/api/v1/projects/my-company/staging-env/initialize"},
		{"org123", "project456", "/api/v1/projects/org123/project456/initialize"},
	}

	for _, tc := range testCases {
		t.Run("project_"+tc.orgSlug+"_"+tc.projectSlug, func(t *testing.T) {
			route := Project.InitializeKey(tc.orgSlug, tc.projectSlug)
			assert.Equal(t, tc.expected, route)
		})
	}
}

func TestDeviceRoutesWithDifferentIDs(t *testing.T) {
	// Test with various device ID formats
	testCases := []struct {
		deviceID string
		expected string
	}{
		{"abc123", "/api/v1/devices/abc123"},
		{"device-uuid-12345", "/api/v1/devices/device-uuid-12345"},
		{"DEV_001", "/api/v1/devices/DEV_001"},
	}

	for _, tc := range testCases {
		t.Run("device_"+tc.deviceID, func(t *testing.T) {
			route := Device.GetByID(tc.deviceID)
			assert.Equal(t, tc.expected, route)
		})
	}
}
