package client

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/encoding"
	"github.com/InitiatDev/initiat-cli/internal/httputil"
	"github.com/InitiatDev/initiat-cli/internal/routes"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

const (
	defaultTimeoutSeconds = 30
	debugPreviewLength    = 20 // Length of key preview for debug output
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New() *Client {
	cfg := config.Get()

	return &Client{
		baseURL: cfg.API.BaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeoutSeconds * time.Second,
		},
	}
}

func NewWithBaseURL(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeoutSeconds * time.Second,
		},
	}
}

func (c *Client) Login(email, password string) (*types.LoginResponse, error) {
	loginReq := types.LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.AuthLogin)
	statusCode, body, err := httputil.DoUnsignedRequest(c.httpClient, routes.POST, url, jsonData)
	if err != nil {
		return nil, err
	}

	var loginResp types.LoginResponse
	if err := httputil.HandleGetResponse(statusCode, body, &loginResp); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &loginResp, nil
}

func (c *Client) encodeKeys(signingPublicKey ed25519.PublicKey, encryptionPublicKey []byte) (string, string, error) {
	ed25519Encoded, err := encoding.EncodeEd25519PublicKey(signingPublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode Ed25519 public key: %w", err)
	}

	x25519Encoded, err := encoding.EncodeX25519PublicKey(encryptionPublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode X25519 public key: %w", err)
	}

	// Debug: show encoded keys (first few chars only for security)
	ed25519Preview := ed25519Encoded
	if len(ed25519Preview) > debugPreviewLength {
		ed25519Preview = ed25519Preview[:debugPreviewLength] + "..."
	}
	x25519Preview := x25519Encoded
	if len(x25519Preview) > debugPreviewLength {
		x25519Preview = x25519Preview[:debugPreviewLength] + "..."
	}
	fmt.Printf("🔍 Debug: Ed25519 encoded: %s\n", ed25519Preview)
	fmt.Printf("🔍 Debug: X25519 encoded: %s\n", x25519Preview)

	return ed25519Encoded, x25519Encoded, nil
}

func (c *Client) RegisterDevice(
	token, name string,
	signingPublicKey ed25519.PublicKey,
	encryptionPublicKey []byte,
) (*types.DeviceRegistrationResponse, error) {
	ed25519Encoded, x25519Encoded, err := c.encodeKeys(signingPublicKey, encryptionPublicKey)
	if err != nil {
		return nil, err
	}

	deviceReq := types.DeviceRegistrationRequest{
		Token:            token,
		Name:             name,
		PublicKeyEd25519: ed25519Encoded,
		PublicKeyX25519:  x25519Encoded,
	}

	jsonData, err := json.Marshal(deviceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device registration request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.Devices)
	statusCode, body, err := httputil.DoUnsignedRequest(c.httpClient, routes.POST, url, jsonData)
	if err != nil {
		return nil, err
	}

	var deviceResp types.DeviceRegistrationResponse
	if err := httputil.HandleStandardResponse(statusCode, body, &deviceResp); err != nil {
		return nil, fmt.Errorf("device registration failed: %w", err)
	}

	return &deviceResp, nil
}

func (c *Client) ListWorkspaces() ([]types.Workspace, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspaces)
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return nil, err
	}

	var workspacesResp types.ListWorkspacesResponse
	if err := httputil.HandleGetResponse(statusCode, body, &workspacesResp); err != nil {
		return nil, fmt.Errorf("list workspaces failed: %w", err)
	}

	return workspacesResp.Workspaces, nil
}

func (c *Client) GetWorkspaceBySlug(orgSlug, workspaceSlug string) (*types.Workspace, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.GetBySlug(orgSlug, workspaceSlug))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return nil, err
	}

	var workspace types.Workspace
	if err := httputil.HandleGetResponse(statusCode, body, &workspace); err != nil {
		return nil, fmt.Errorf("get workspace failed: %w", err)
	}

	return &workspace, nil
}

func (c *Client) InitializeWorkspaceKey(orgSlug, workspaceSlug string, wrappedKey []byte) error {
	initReq := types.InitializeWorkspaceKeyRequest{
		WrappedWorkspaceKey: encoding.Encode(wrappedKey),
	}

	jsonData, err := json.Marshal(initReq)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize key request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.Workspace.InitializeKey(orgSlug, workspaceSlug))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.POST, url, jsonData)
	if err != nil {
		return err
	}

	return httputil.HandleStandardResponse(statusCode, body, nil)
}

func (c *Client) GetWrappedWorkspaceKey(orgSlug, workspaceSlug string) (string, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.GetWorkspaceKey(orgSlug, workspaceSlug))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return "", err
	}

	var keyResp types.GetWorkspaceKeyResponse
	if err := httputil.HandleGetResponse(statusCode, body, &keyResp); err != nil {
		return "", fmt.Errorf("get workspace key failed: %w", err)
	}

	return keyResp.WrappedWorkspaceKey, nil
}

func (c *Client) SetSecret(
	orgSlug, workspaceSlug, key string, encryptedValue, nonce []byte, description string, force bool,
) (*types.Secret, error) {
	setReq := types.SetSecretRequest{
		Key:            key,
		EncryptedValue: encoding.Encode(encryptedValue),
		Nonce:          encoding.Encode(nonce),
		Description:    description,
	}

	jsonData, err := json.Marshal(setReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal set secret request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.Workspace.Secrets(orgSlug, workspaceSlug))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.POST, url, jsonData)
	if err != nil {
		return nil, err
	}

	var setResp types.SetSecretResponse
	if err := httputil.HandleStandardResponse(statusCode, body, &setResp); err != nil {
		return nil, fmt.Errorf("set secret failed: %w", err)
	}

	return &setResp.Secret, nil
}

func (c *Client) GetSecret(orgSlug, workspaceSlug, secretKey string) (*types.SecretWithValue, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.SecretByKey(orgSlug, workspaceSlug, secretKey))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return nil, err
	}

	var getResp types.GetSecretResponse
	if err := httputil.HandleGetResponse(statusCode, body, &getResp); err != nil {
		return nil, fmt.Errorf("get secret failed: %w", err)
	}

	return &getResp.Secret, nil
}

func (c *Client) ListSecrets(orgSlug, workspaceSlug string) ([]types.Secret, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.Secrets(orgSlug, workspaceSlug))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return nil, err
	}

	var listResp types.ListSecretsResponse
	if err := httputil.HandleGetResponse(statusCode, body, &listResp); err != nil {
		return nil, fmt.Errorf("list secrets failed: %w", err)
	}

	return listResp.Secrets, nil
}

func (c *Client) DeleteSecret(orgSlug, workspaceSlug, secretKey string) error {
	url := routes.BuildURL(c.baseURL, routes.Workspace.SecretByKey(orgSlug, workspaceSlug, secretKey))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.DELETE, url, nil)
	if err != nil {
		return err
	}

	return httputil.HandleDeleteResponse(statusCode, body)
}

func (c *Client) ListDeviceApprovals() ([]types.DeviceApproval, error) {
	url := routes.BuildURL(c.baseURL, routes.DeviceApprovals)
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return nil, err
	}

	var approvalsResp types.ListDeviceApprovalsResponse
	if err := httputil.HandleGetResponse(statusCode, body, &approvalsResp); err != nil {
		return nil, fmt.Errorf("list device approvals failed: %w", err)
	}

	return approvalsResp.DeviceApprovals, nil
}

func (c *Client) GetDeviceApproval(approvalID string) (*types.DeviceApproval, error) {
	url := routes.BuildURL(c.baseURL, routes.DeviceApproval.GetByID(approvalID))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.GET, url, nil)
	if err != nil {
		return nil, err
	}

	var approvalResp types.GetDeviceApprovalResponse
	if err := httputil.HandleGetResponse(statusCode, body, &approvalResp); err != nil {
		return nil, fmt.Errorf("get device approval failed: %w", err)
	}

	return &approvalResp.DeviceApproval, nil
}

func (c *Client) ApproveDevice(approvalID string, wrappedWorkspaceKey string) (*types.DeviceApproval, error) {
	approveReq := types.ApproveDeviceRequest{
		WrappedWorkspaceKey: wrappedWorkspaceKey,
	}

	jsonData, err := json.Marshal(approveReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approve device request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.DeviceApproval.Approve(approvalID))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.POST, url, jsonData)
	if err != nil {
		return nil, err
	}

	var approveResp types.ApproveDeviceResponse
	if err := httputil.HandleStandardResponse(statusCode, body, &approveResp); err != nil {
		return nil, fmt.Errorf("approve device failed: %w", err)
	}

	return &approveResp.DeviceApproval, nil
}

func (c *Client) RejectDevice(approvalID string) (*types.DeviceApproval, error) {
	url := routes.BuildURL(c.baseURL, routes.DeviceApproval.Reject(approvalID))
	statusCode, body, err := httputil.DoSignedRequest(c.httpClient, routes.POST, url, nil)
	if err != nil {
		return nil, err
	}

	var rejectResp types.RejectDeviceResponse
	if err := httputil.HandleStandardResponse(statusCode, body, &rejectResp); err != nil {
		return nil, fmt.Errorf("reject device failed: %w", err)
	}

	return &rejectResp.DeviceApproval, nil
}
