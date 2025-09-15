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

func TestWorkspaceRoutes(t *testing.T) {
	assert.Equal(t, "/api/v1/workspaces", Workspaces)
}

func TestWorkspaceRoutes_InitializeKey(t *testing.T) {
	route := Workspace.InitializeKey(123)
	assert.Equal(t, "/api/v1/workspaces/123/initialize", route)
}

func TestWorkspaceRoutes_GetByID(t *testing.T) {
	route := Workspace.GetByID(456)
	assert.Equal(t, "/api/v1/workspaces/456", route)
}

func TestWorkspaceRoutes_Secrets(t *testing.T) {
	route := Workspace.Secrets(789)
	assert.Equal(t, "/api/v1/workspaces/789/secrets", route)
}

func TestWorkspaceRoutes_SecretByKey(t *testing.T) {
	route := Workspace.SecretByKey(123, "API_KEY")
	assert.Equal(t, "/api/v1/workspaces/123/secrets/API_KEY", route)
}

func TestWorkspaceRoutes_InviteDevice(t *testing.T) {
	route := Workspace.InviteDevice(456)
	assert.Equal(t, "/api/v1/workspaces/456/invite-device", route)
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
	assert.Equal(t, Route{Method: GET, Path: "/api/v1/workspaces"}, ListWorkspacesRoute)
	assert.Equal(t, Route{Method: POST, Path: "/api/v1/devices"}, RegisterDeviceRoute)
	assert.Equal(t, Route{Method: GET, Path: "/api/v1/devices"}, ListDevicesRoute)
}

func TestBuildURL(t *testing.T) {
	baseURL := "https://api.initflow.com"

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
			expected: "https://api.initflow.com/api/v1/auth/login",
		},
		{
			name:     "workspaces route",
			baseURL:  baseURL,
			path:     Workspaces,
			expected: "https://api.initflow.com/api/v1/workspaces",
		},
		{
			name:     "localhost development",
			baseURL:  "http://localhost:4000",
			path:     AuthLogin,
			expected: "http://localhost:4000/api/v1/auth/login",
		},
		{
			name:     "workspace initialize key",
			baseURL:  baseURL,
			path:     Workspace.InitializeKey(123),
			expected: "https://api.initflow.com/api/v1/workspaces/123/initialize",
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

func TestWorkspaceRoutesWithDifferentIDs(t *testing.T) {
	// Test with various workspace IDs to ensure proper formatting
	testCases := []struct {
		workspaceID int
		expected    string
	}{
		{1, "/api/v1/workspaces/1/initialize"},
		{999, "/api/v1/workspaces/999/initialize"},
		{123456, "/api/v1/workspaces/123456/initialize"},
	}

	for _, tc := range testCases {
		t.Run("workspace_"+string(rune(tc.workspaceID)), func(t *testing.T) {
			route := Workspace.InitializeKey(tc.workspaceID)
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
