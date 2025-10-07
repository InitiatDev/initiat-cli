package routes

import "fmt"

const APIVersion = "v1"

const (
	APIBasePath = "/api/" + APIVersion
)

const (
	AuthLogin       = APIBasePath + "/auth/login"
	Devices         = APIBasePath + "/devices"
	Projects        = APIBasePath + "/projects"
	DeviceApprovals = APIBasePath + "/device-approvals"
)

type ProjectRoutes struct{}

func (w ProjectRoutes) InitializeKey(orgSlug, projectSlug string) string {
	return fmt.Sprintf("%s/%s/%s/initialize", Projects, orgSlug, projectSlug)
}

func (w ProjectRoutes) GetBySlug(orgSlug, projectSlug string) string {
	return fmt.Sprintf("%s/%s/%s", Projects, orgSlug, projectSlug)
}

func (w ProjectRoutes) Secrets(orgSlug, projectSlug string) string {
	return fmt.Sprintf("%s/%s/%s/secrets", Projects, orgSlug, projectSlug)
}

func (w ProjectRoutes) SecretByKey(orgSlug, projectSlug, secretKey string) string {
	return fmt.Sprintf("%s/%s/%s/secrets/%s", Projects, orgSlug, projectSlug, secretKey)
}

func (w ProjectRoutes) InviteDevice(orgSlug, projectSlug string) string {
	return fmt.Sprintf("%s/%s/%s/invite-device", Projects, orgSlug, projectSlug)
}

func (w ProjectRoutes) GetProjectKey(orgSlug, projectSlug string) string {
	return fmt.Sprintf("%s/%s/%s/project_key", Projects, orgSlug, projectSlug)
}

var Project = ProjectRoutes{}

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

type DeviceApprovalRoutes struct{}

func (d DeviceApprovalRoutes) GetByID(approvalID string) string {
	return fmt.Sprintf("%s/%s", DeviceApprovals, approvalID)
}

func (d DeviceApprovalRoutes) Approve(approvalID string) string {
	return fmt.Sprintf("%s/%s/approve", DeviceApprovals, approvalID)
}

func (d DeviceApprovalRoutes) Reject(approvalID string) string {
	return fmt.Sprintf("%s/%s/reject", DeviceApprovals, approvalID)
}

var DeviceApproval = DeviceApprovalRoutes{}

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

	ListProjectsRoute = Route{
		Method: GET,
		Path:   Projects,
	}

	RegisterDeviceRoute = Route{
		Method: POST,
		Path:   Devices,
	}

	ListDevicesRoute = Route{
		Method: GET,
		Path:   Devices,
	}

	ListDeviceApprovalsRoute = Route{
		Method: GET,
		Path:   DeviceApprovals,
	}
)

func BuildURL(baseURL, routePath string) string {
	return baseURL + routePath
}
