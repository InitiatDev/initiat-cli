package routes

import "fmt"

const APIVersion = "v1"

const (
	APIBasePath = "/api/" + APIVersion
)

const (
	AuthLogin  = APIBasePath + "/auth/login"
	Devices    = APIBasePath + "/devices"
	Workspaces = APIBasePath + "/workspaces"
)

type WorkspaceRoutes struct{}

func (w WorkspaceRoutes) InitializeKey(workspaceID int) string {
	return fmt.Sprintf("%s/%d/initialize-key", Workspaces, workspaceID)
}

func (w WorkspaceRoutes) GetByID(workspaceID int) string {
	return fmt.Sprintf("%s/%d", Workspaces, workspaceID)
}

func (w WorkspaceRoutes) Secrets(workspaceID int) string {
	return fmt.Sprintf("%s/%d/secrets", Workspaces, workspaceID)
}

func (w WorkspaceRoutes) SecretByKey(workspaceID int, secretKey string) string {
	return fmt.Sprintf("%s/%d/secrets/%s", Workspaces, workspaceID, secretKey)
}

func (w WorkspaceRoutes) InviteDevice(workspaceID int) string {
	return fmt.Sprintf("%s/%d/invite-device", Workspaces, workspaceID)
}

var Workspace = WorkspaceRoutes{}

type DeviceRoutes struct{}

func (d DeviceRoutes) GetByID(deviceID string) string {
	return fmt.Sprintf("%s/%s", Devices, deviceID)
}

func (d DeviceRoutes) Revoke(deviceID string) string {
	return fmt.Sprintf("%s/%s/revoke", Devices, deviceID)
}

var Device = DeviceRoutes{}

type SecretRoutes struct{}

var Secret = SecretRoutes{}

type Route struct {
	Method string
	Path   string
}

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	PATCH  = "PATCH"
)

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

func BuildURL(baseURL, routePath string) string {
	return baseURL + routePath
}
