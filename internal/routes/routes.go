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

func (w WorkspaceRoutes) InitializeKey(orgSlug, workspaceSlug string) string {
	return fmt.Sprintf("%s/%s/%s/initialize", Workspaces, orgSlug, workspaceSlug)
}

func (w WorkspaceRoutes) GetBySlug(orgSlug, workspaceSlug string) string {
	return fmt.Sprintf("%s/%s/%s", Workspaces, orgSlug, workspaceSlug)
}

func (w WorkspaceRoutes) Secrets(orgSlug, workspaceSlug string) string {
	return fmt.Sprintf("%s/%s/%s/secrets", Workspaces, orgSlug, workspaceSlug)
}

func (w WorkspaceRoutes) SecretByKey(orgSlug, workspaceSlug, secretKey string) string {
	return fmt.Sprintf("%s/%s/%s/secrets/%s", Workspaces, orgSlug, workspaceSlug, secretKey)
}

func (w WorkspaceRoutes) InviteDevice(orgSlug, workspaceSlug string) string {
	return fmt.Sprintf("%s/%s/%s/invite-device", Workspaces, orgSlug, workspaceSlug)
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
