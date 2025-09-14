package routes

import "fmt"

// APIVersion represents the API version
const APIVersion = "v1"

// Base paths
const (
	APIBasePath = "/api/" + APIVersion
)

// Authentication routes
const (
	AuthLogin = APIBasePath + "/auth/login"
)

// Device routes
const (
	Devices = APIBasePath + "/devices"
)

// Workspace routes
const (
	Workspaces = APIBasePath + "/workspaces"
)

// WorkspaceRoutes provides methods for workspace-specific routes
type WorkspaceRoutes struct{}

// InitializeKey returns the route for initializing a workspace key
func (w WorkspaceRoutes) InitializeKey(workspaceID int) string {
	return fmt.Sprintf("%s/%d/initialize-key", Workspaces, workspaceID)
}

// GetByID returns the route for getting a specific workspace
func (w WorkspaceRoutes) GetByID(workspaceID int) string {
	return fmt.Sprintf("%s/%d", Workspaces, workspaceID)
}

// Secrets returns the route for workspace secrets
func (w WorkspaceRoutes) Secrets(workspaceID int) string {
	return fmt.Sprintf("%s/%d/secrets", Workspaces, workspaceID)
}

// SecretByKey returns the route for a specific secret
func (w WorkspaceRoutes) SecretByKey(workspaceID int, secretKey string) string {
	return fmt.Sprintf("%s/%d/secrets/%s", Workspaces, workspaceID, secretKey)
}

// InviteDevice returns the route for inviting a device to a workspace
func (w WorkspaceRoutes) InviteDevice(workspaceID int) string {
	return fmt.Sprintf("%s/%d/invite-device", Workspaces, workspaceID)
}

// Workspace provides workspace-specific route methods
var Workspace = WorkspaceRoutes{}

// DeviceRoutes provides methods for device-specific routes
type DeviceRoutes struct{}

// GetByID returns the route for getting a specific device
func (d DeviceRoutes) GetByID(deviceID string) string {
	return fmt.Sprintf("%s/%s", Devices, deviceID)
}

// Revoke returns the route for revoking a device
func (d DeviceRoutes) Revoke(deviceID string) string {
	return fmt.Sprintf("%s/%s/revoke", Devices, deviceID)
}

// Device provides device-specific route methods
var Device = DeviceRoutes{}

// SecretRoutes provides methods for secret-specific routes (future use)
type SecretRoutes struct{}

// Secret provides secret-specific route methods
var Secret = SecretRoutes{}

// Route represents a complete route with method and path
type Route struct {
	Method string
	Path   string
}

// Common HTTP methods
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	PATCH  = "PATCH"
)

// Predefined routes for common operations
var (
	LoginRoute = Route{
		Method: POST,
		Path:   AuthLogin,
	}

	ListWorkspacesRoute = Route{
		Method: GET,
		Path:   Workspaces,
	}

	RegisterDeviceRoute = Route{
		Method: POST,
		Path:   Devices,
	}

	ListDevicesRoute = Route{
		Method: GET,
		Path:   Devices,
	}
)

// BuildURL constructs a full URL from base URL and route path
func BuildURL(baseURL, routePath string) string {
	return baseURL + routePath
}
