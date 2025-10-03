package client

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/InitiatDev/initiat-cli/internal/routes"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func TestNew(t *testing.T) {
	client := New()
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)

	assert.NotEmpty(t, client.baseURL)
}

func TestNewWithBaseURL(t *testing.T) {
	baseURL := "http://localhost:4000"
	client := NewWithBaseURL(baseURL)
	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestLogin_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, routes.POST, r.Method)
		assert.Equal(t, routes.AuthLogin, r.URL.Path)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "initiat-cli/1.0", r.Header.Get("User-Agent"))

		var loginReq types.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&loginReq)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", loginReq.Email)
		assert.Equal(t, "password123", loginReq.Password)

		loginData := types.LoginResponse{
			Token: "test-token-123",
			User: struct {
				ID      int    `json:"id"`
				Email   string `json:"email"`
				Name    string `json:"name"`
				Surname string `json:"surname"`
			}{
				ID:      1,
				Email:   "test@example.com",
				Name:    "John",
				Surname: "Doe",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Authentication successful",
			"data":    loginData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	resp, err := client.Login("test@example.com", "password123")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-token-123", resp.Token)
	assert.Equal(t, 1, resp.User.ID)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "John", resp.User.Name)
	assert.Equal(t, "Doe", resp.User.Surname)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": false,
			"message": "Invalid email or password",
			"errors":  []string{"Invalid email or password"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	resp, err := client.Login("test@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid email or password")
}

func TestLogin_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse error response")
}

func TestLogin_NetworkError(t *testing.T) {
	client := NewWithBaseURL("http://192.0.2.0:1")

	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)

	expectedErrors := []string{
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"dial tcp 192.0.2.0:1: connect: connection refused",
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"dial tcp 192.0.2.0:1: i/o timeout",
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"dial tcp 192.0.2.0:1: i/o timeout (Client.Timeout exceeded while awaiting headers)",
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"dial tcp 192.0.2.0:1: network is unreachable",
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"context deadline exceeded (Client.Timeout exceeded while awaiting headers)",
	}

	errorMatched := false
	for _, expectedError := range expectedErrors {
		if err.Error() == expectedError {
			errorMatched = true
			break
		}
	}
	assert.True(t, errorMatched, "Expected a network error, got: %s", err.Error())
}

func TestLogin_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse API response")
}

func TestLogin_EmptyCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loginData := types.LoginResponse{
			Token: "test-token",
			User: struct {
				ID      int    `json:"id"`
				Email   string `json:"email"`
				Name    string `json:"name"`
				Surname string `json:"surname"`
			}{
				ID:      1,
				Email:   "test@example.com",
				Name:    "Test",
				Surname: "User",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Authentication successful",
			"data":    loginData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	_, err := client.Login("", "password123")
	assert.NoError(t, err)

	_, err = client.Login("test@example.com", "")
	assert.NoError(t, err)
}

func TestRegisterDevice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, routes.Devices, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "initiat-cli/1.0", r.Header.Get("User-Agent"))

		var req types.DeviceRegistrationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-token", req.Token)
		assert.Equal(t, "Test Device", req.Name)
		assert.NotEmpty(t, req.PublicKeyEd25519)
		assert.NotEmpty(t, req.PublicKeyX25519)

		deviceData := types.DeviceRegistrationResponse{
			Success: true,
			Message: "Device registered successfully",
			Device: struct {
				DeviceID  string `json:"device_id"`
				Name      string `json:"name"`
				CreatedAt string `json:"created_at"`
			}{
				DeviceID:  "device-123",
				Name:      "Test Device",
				CreatedAt: "2025-09-13T14:30:22Z",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Device registered successfully",
			"data":    deviceData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	resp, err := client.RegisterDevice("test-token", "Test Device", signingPublic, encryptionPublic)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Device registered successfully", resp.Message)
	assert.Equal(t, "device-123", resp.Device.DeviceID)
	assert.Equal(t, "Test Device", resp.Device.Name)
	assert.Equal(t, "2025-09-13T14:30:22Z", resp.Device.CreatedAt)
}

func TestRegisterDevice_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": false,
			"message": "Invalid device name",
			"errors":  []string{"Invalid device name"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	_, err = client.RegisterDevice("test-token", "", signingPublic, encryptionPublic)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid device name")
}

func TestRegisterDevice_NetworkError(t *testing.T) {
	client := NewWithBaseURL("http://invalid-url-that-does-not-exist.local")

	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	_, err = client.RegisterDevice("test-token", "Test Device", signingPublic, encryptionPublic)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
}

func TestRegisterDevice_Success_With200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceData := types.DeviceRegistrationResponse{
			Success: true,
			Message: "Device registered successfully",
			Device: struct {
				DeviceID  string `json:"device_id"`
				Name      string `json:"name"`
				CreatedAt string `json:"created_at"`
			}{
				DeviceID:  "device-456",
				Name:      "Test Device 200",
				CreatedAt: "2025-09-14T14:30:22Z",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Device registered successfully",
			"data":    deviceData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	resp, err := client.RegisterDevice("test-token", "Test Device 200", signingPublic, encryptionPublic)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "device-456", resp.Device.DeviceID)
}
