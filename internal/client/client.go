package client

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DylanBlakemore/initflow-cli/internal/config"
	"github.com/DylanBlakemore/initflow-cli/internal/encoding"
	"github.com/DylanBlakemore/initflow-cli/internal/routes"
	"github.com/DylanBlakemore/initflow-cli/internal/storage"
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
		baseURL: cfg.APIBaseURL,
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

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
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

type Workspace struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
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

type ListWorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}

type InitializeWorkspaceKeyRequest struct {
	WrappedWorkspaceKey string `json:"wrapped_workspace_key"`
}

type InitializeWorkspaceKeyResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Workspace Workspace `json:"workspace"`
}

func (c *Client) Login(email, password string) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.AuthLogin)
	req, err := http.NewRequest(routes.POST, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "initflow-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("login failed: %s", errResp.Message)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
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

func (c *Client) handleRegistrationResponse(resp *http.Response, body []byte) (*DeviceRegistrationResponse, error) {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("device registration failed with status %d, raw response: %s", resp.StatusCode, string(body))
		}
		if errResp.Message == "" {
			return nil, fmt.Errorf("device registration failed with status %d, error: %s, raw response: %s",
				resp.StatusCode, errResp.Error, string(body))
		}
		return nil, fmt.Errorf("device registration failed: %s", errResp.Message)
	}

	var deviceResp DeviceRegistrationResponse
	if err := json.Unmarshal(body, &deviceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &deviceResp, nil
}

func (c *Client) RegisterDevice(
	token, name string,
	signingPublicKey ed25519.PublicKey,
	encryptionPublicKey []byte,
) (*DeviceRegistrationResponse, error) {
	ed25519Encoded, x25519Encoded, err := c.encodeKeys(signingPublicKey, encryptionPublicKey)
	if err != nil {
		return nil, err
	}

	deviceReq := DeviceRegistrationRequest{
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
	req, err := http.NewRequest(routes.POST, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "initflow-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return c.handleRegistrationResponse(resp, body)
}

func (c *Client) signRequest(req *http.Request, body []byte) error {
	store := storage.New()

	deviceID, err := store.GetDeviceID()
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	signingKey, err := store.GetSigningPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}

	timestamp := time.Now().Unix()

	var bodyStr string
	if body != nil {
		bodyStr = string(body)
	}

	message := fmt.Sprintf("%s\n%s\n%s\n%d",
		req.Method,
		req.URL.Path+req.URL.RawQuery,
		bodyStr,
		timestamp)

	signature := ed25519.Sign(signingKey, []byte(message))

	signatureEncoded, err := encoding.EncodeEd25519Signature(signature)
	if err != nil {
		return fmt.Errorf("failed to encode signature: %w", err)
	}

	req.Header.Set("Authorization", "Device "+deviceID)
	req.Header.Set("X-Signature", signatureEncoded)
	req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))

	return nil
}

func (c *Client) ListWorkspaces() ([]Workspace, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspaces)
	req, err := http.NewRequest(routes.GET, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "initflow-cli/1.0")

	if err := c.signRequest(req, nil); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("list workspaces failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("list workspaces failed: %s", errResp.Message)
	}

	var workspacesResp ListWorkspacesResponse
	if err := json.Unmarshal(body, &workspacesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return workspacesResp.Workspaces, nil
}

func (c *Client) GetWorkspaceBySlug(slug string) (*Workspace, error) {
	workspaces, err := c.ListWorkspaces()
	if err != nil {
		return nil, err
	}

	for _, workspace := range workspaces {
		if workspace.Slug == slug {
			return &workspace, nil
		}
	}

	return nil, fmt.Errorf("workspace '%s' not found", slug)
}

func (c *Client) InitializeWorkspaceKey(workspaceID int, wrappedKey []byte) error {
	initReq := InitializeWorkspaceKeyRequest{
		WrappedWorkspaceKey: encoding.Encode(wrappedKey),
	}

	jsonData, err := json.Marshal(initReq)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize key request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.Workspace.InitializeKey(workspaceID))
	req, err := http.NewRequest(routes.POST, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "initflow-cli/1.0")

	if err := c.signRequest(req, jsonData); err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("initialize workspace key failed with status %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("initialize workspace key failed: %s", errResp.Message)
	}

	return nil
}
