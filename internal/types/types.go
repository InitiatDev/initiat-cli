package types

import "encoding/json"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID      int    `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Surname string `json:"surname"`
	} `json:"user"`
}

type DeviceRegistrationRequest struct {
	Token            string `json:"token"`
	Name             string `json:"name"`
	PublicKeyEd25519 string `json:"public_key_ed25519"`
	PublicKeyX25519  string `json:"public_key_x25519"`
}

type DeviceRegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Device  struct {
		DeviceID  string `json:"device_id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	} `json:"device"`
}

type Project struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	CompositeSlug  string `json:"composite_slug"`
	Description    string `json:"description"`
	KeyInitialized bool   `json:"key_initialized"`
	KeyVersion     int    `json:"key_version"`
	Role           string `json:"role"`
	Organization   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"organization"`
}

type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type InitializeProjectKeyRequest struct {
	WrappedProjectKey string `json:"wrapped_project_key"`
}

type InitializeProjectKeyResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Project Project `json:"project"`
}

type GetProjectKeyResponse struct {
	WrappedProjectKey string `json:"wrapped_project_key"`
	KeyVersion        int    `json:"key_version"`
}

type Secret struct {
	ID              int    `json:"id"`
	Key             string `json:"key"`
	Version         int    `json:"version"`
	ProjectID       int    `json:"project_id"`
	CreatedByDevice struct {
		ID       int    `json:"id"`
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
	} `json:"created_by_device"`
	InsertedAt string `json:"inserted_at"`
	UpdatedAt  string `json:"updated_at"`
}

type SecretWithValue struct {
	Secret
	EncryptedValue string `json:"encrypted_value"`
	Nonce          string `json:"nonce"`
}

type SetSecretRequest struct {
	Key            string `json:"key"`
	EncryptedValue string `json:"encrypted_value"`
	Nonce          string `json:"nonce"`
	Description    string `json:"description,omitempty"`
}

type SetSecretResponse struct {
	Secret Secret `json:"secret"`
}

type GetSecretResponse struct {
	Secret SecretWithValue `json:"secret"`
}

type ListSecretsResponse struct {
	Secrets []Secret `json:"secrets"`
	Count   int      `json:"count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Errors  []string        `json:"errors,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type ValidationErrorResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

type DeviceApproval struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	InsertedAt string `json:"inserted_at"`
	UpdatedAt  string `json:"updated_at"`
	Device     struct {
		ID               int    `json:"id"`
		Name             string `json:"name"`
		PublicKeyEd25519 string `json:"public_key_ed25519"`
		PublicKeyX25519  string `json:"public_key_x25519"`
	} `json:"device"`
	ProjectMembership struct {
		ID     int    `json:"id"`
		Role   string `json:"role"`
		Status string `json:"status"`
		User   struct {
			ID      int    `json:"id"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Surname string `json:"surname"`
		} `json:"user"`
		Project struct {
			ID           int    `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			Organization struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"organization"`
		} `json:"project"`
	} `json:"project_membership"`
	ApprovedByUser *struct {
		ID      int    `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Surname string `json:"surname"`
	} `json:"approved_by_user"`
}

type ListDeviceApprovalsResponse struct {
	DeviceApprovals []DeviceApproval `json:"device_approvals"`
}

type GetDeviceApprovalResponse struct {
	DeviceApproval DeviceApproval `json:"device_approval"`
}

type ApproveDeviceRequest struct {
	WrappedProjectKey string `json:"wrapped_project_key"`
}

type ApproveDeviceResponse struct {
	DeviceApproval DeviceApproval `json:"device_approval"`
}

type RejectDeviceResponse struct {
	DeviceApproval DeviceApproval `json:"device_approval"`
}
