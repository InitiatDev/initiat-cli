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

	"github.com/DylanBlakemore/initiat-cli/internal/config"
	"github.com/DylanBlakemore/initiat-cli/internal/encoding"
	"github.com/DylanBlakemore/initiat-cli/internal/routes"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
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

func parseAPIResponse(body []byte, target interface{}) error {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return fmt.Errorf("API error: %s", apiResp.Errors[0])
		}
		return fmt.Errorf("API error: %s", apiResp.Message)
	}

	if target != nil && len(apiResp.Data) > 0 {
		if err := json.Unmarshal(apiResp.Data, target); err != nil {
			return fmt.Errorf("failed to parse response data: %w", err)
		}
	}

	return nil
}

func parseValidationErrorResponse(body []byte) error {
	var validationResp ValidationErrorResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return fmt.Errorf("failed to parse error response: %w", err)
		}
		if len(apiResp.Errors) > 0 {
			return fmt.Errorf("validation error: %s", apiResp.Errors[0])
		}
		return fmt.Errorf("validation error: %s", apiResp.Message)
	}

	if validationResp.Success {
		return nil
	}
	if len(validationResp.Errors) > 0 {
		var errorMessages []string
		for field, messages := range validationResp.Errors {
			for _, msg := range messages {
				errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, msg))
			}
		}
		if len(errorMessages) > 0 {
			return fmt.Errorf("validation failed: %s", errorMessages[0])
		}
	}

	return fmt.Errorf("validation error: %s", validationResp.Message)
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

type Secret struct {
	ID              int    `json:"id"`
	Key             string `json:"key"`
	Version         int    `json:"version"`
	WorkspaceID     int    `json:"workspace_id"`
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
	req.Header.Set("User-Agent", "initiat-cli/1.0")

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
		return nil, parseValidationErrorResponse(body)
	}

	var loginResp LoginResponse
	if err := parseAPIResponse(body, &loginResp); err != nil {
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

func (c *Client) handleRegistrationResponse(resp *http.Response, body []byte) (*DeviceRegistrationResponse, error) {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, parseValidationErrorResponse(body)
	}

	var deviceResp DeviceRegistrationResponse
	if err := parseAPIResponse(body, &deviceResp); err != nil {
		return nil, fmt.Errorf("device registration failed: %w", err)
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
	req.Header.Set("User-Agent", "initiat-cli/1.0")

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

func (c *Client) signRequest(req *http.Request, _ []byte) error {
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

	// ✅ NEW: Body-agnostic signature format (no body in message)
	message := fmt.Sprintf("%s\n%s\n%d",
		req.Method,
		req.URL.Path+req.URL.RawQuery,
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

	req.Header.Set("User-Agent", "initiat-cli/1.0")

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
		return nil, parseValidationErrorResponse(body)
	}

	var workspacesResp ListWorkspacesResponse
	if err := parseAPIResponse(body, &workspacesResp); err != nil {
		return nil, fmt.Errorf("list workspaces failed: %w", err)
	}

	return workspacesResp.Workspaces, nil
}

func (c *Client) GetWorkspaceBySlug(orgSlug, workspaceSlug string) (*Workspace, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.GetBySlug(orgSlug, workspaceSlug))
	req, err := http.NewRequest(routes.GET, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "initiat-cli/1.0")

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
		return nil, parseValidationErrorResponse(body)
	}

	var workspace Workspace
	if err := parseAPIResponse(body, &workspace); err != nil {
		return nil, fmt.Errorf("get workspace failed: %w", err)
	}

	return &workspace, nil
}

func (c *Client) InitializeWorkspaceKey(orgSlug, workspaceSlug string, wrappedKey []byte) error {
	initReq := InitializeWorkspaceKeyRequest{
		WrappedWorkspaceKey: encoding.Encode(wrappedKey),
	}

	jsonData, err := json.Marshal(initReq)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize key request: %w", err)
	}

	url := routes.BuildURL(c.baseURL, routes.Workspace.InitializeKey(orgSlug, workspaceSlug))
	req, err := http.NewRequest(routes.POST, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "initiat-cli/1.0")

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
		return parseValidationErrorResponse(body)
	}

	return parseAPIResponse(body, nil)
}

func (c *Client) SetSecret(
	orgSlug, workspaceSlug, key string, encryptedValue, nonce []byte, description string, force bool,
) (*Secret, error) {
	setReq := SetSecretRequest{
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
	req, err := http.NewRequest(routes.POST, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "initiat-cli/1.0")

	if err := c.signRequest(req, jsonData); err != nil {
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, parseValidationErrorResponse(body)
	}

	var setResp SetSecretResponse
	if err := parseAPIResponse(body, &setResp); err != nil {
		return nil, fmt.Errorf("set secret failed: %w", err)
	}

	return &setResp.Secret, nil
}

func (c *Client) GetSecret(orgSlug, workspaceSlug, secretKey string) (*SecretWithValue, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.SecretByKey(orgSlug, workspaceSlug, secretKey))
	req, err := http.NewRequest(routes.GET, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "initiat-cli/1.0")

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
		return nil, parseValidationErrorResponse(body)
	}

	var getResp GetSecretResponse
	if err := parseAPIResponse(body, &getResp); err != nil {
		return nil, fmt.Errorf("get secret failed: %w", err)
	}

	return &getResp.Secret, nil
}

func (c *Client) ListSecrets(orgSlug, workspaceSlug string) ([]Secret, error) {
	url := routes.BuildURL(c.baseURL, routes.Workspace.Secrets(orgSlug, workspaceSlug))
	req, err := http.NewRequest(routes.GET, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "initiat-cli/1.0")

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
		return nil, parseValidationErrorResponse(body)
	}

	var listResp ListSecretsResponse
	if err := parseAPIResponse(body, &listResp); err != nil {
		return nil, fmt.Errorf("list secrets failed: %w", err)
	}

	return listResp.Secrets, nil
}

func (c *Client) DeleteSecret(orgSlug, workspaceSlug, secretKey string) error {
	url := routes.BuildURL(c.baseURL, routes.Workspace.SecretByKey(orgSlug, workspaceSlug, secretKey))
	req, err := http.NewRequest(routes.DELETE, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "initiat-cli/1.0")

	if err := c.signRequest(req, nil); err != nil {
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseValidationErrorResponse(body)
	}

	return parseAPIResponse(body, nil)
}
