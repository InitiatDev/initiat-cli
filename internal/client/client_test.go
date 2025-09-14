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

	"github.com/DylanBlakemore/initflow-cli/internal/routes"
)

func TestNew(t *testing.T) {
	client := New()
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	// Default base URL should be set from config
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
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, routes.POST, r.Method)
		assert.Equal(t, routes.AuthLogin, r.URL.Path)

		// Verify headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "initflow-cli/1.0", r.Header.Get("User-Agent"))

		// Verify request body
		var loginReq LoginRequest
		err := json.NewDecoder(r.Body).Decode(&loginReq)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", loginReq.Email)
		assert.Equal(t, "password123", loginReq.Password)

		// Return success response
		response := LoginResponse{
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewWithBaseURL(server.URL)

	// Test login
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
	// Create test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid email or password",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewWithBaseURL(server.URL)

	// Test login with invalid credentials
	resp, err := client.Login("test@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid email or password")
}

func TestLogin_ServerError(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewWithBaseURL(server.URL)

	// Test login with server error
	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "login failed with status 500")
}

func TestLogin_NetworkError(t *testing.T) {
	// Create client with invalid URL that will definitely cause a network error
	client := NewWithBaseURL("http://192.0.2.0:1") // RFC5737 test address that should not respond

	// Test login with network error
	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)
	// The error message may vary depending on the system, so just check that it's a network-related error
	expectedErrors := []string{
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"dial tcp 192.0.2.0:1: connect: connection refused",
		"failed to make request: Post \"http://192.0.2.0:1/api/v1/auth/login\": " +
			"dial tcp 192.0.2.0:1: i/o timeout",
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
	// Create test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewWithBaseURL(server.URL)

	// Test login with invalid JSON response
	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestLogin_EmptyCredentials(t *testing.T) {
	// Create test server that accepts any credentials for this test
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LoginResponse{
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	// Test with empty email - client doesn't validate, just sends to server
	_, err := client.Login("", "password123")
	assert.NoError(t, err) // The client doesn't validate empty fields, server would handle this

	// Test with empty password - client doesn't validate, just sends to server
	_, err = client.Login("test@example.com", "")
	assert.NoError(t, err) // The client doesn't validate empty fields, server would handle this
}

func TestRegisterDevice_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, routes.Devices, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "initflow-cli/1.0", r.Header.Get("User-Agent"))

		// Parse request body
		var req DeviceRegistrationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-token", req.Token)
		assert.Equal(t, "Test Device", req.Name)
		assert.NotEmpty(t, req.PublicKeyEd25519)
		assert.NotEmpty(t, req.PublicKeyX25519)

		// Mock successful response
		resp := DeviceRegistrationResponse{
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // Device registration typically returns 201 Created
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	// Generate test keypairs
	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	// Test device registration
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
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid device name",
		})
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	// Generate test keypairs
	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	// Test device registration with server error
	_, err = client.RegisterDevice("test-token", "", signingPublic, encryptionPublic)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid device name")
}

func TestRegisterDevice_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	client := NewWithBaseURL("http://invalid-url-that-does-not-exist.local")

	// Generate test keypairs
	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	// Test device registration with network error
	_, err = client.RegisterDevice("test-token", "Test Device", signingPublic, encryptionPublic)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
}

func TestRegisterDevice_Success_With200(t *testing.T) {
	// Test that we also accept 200 OK responses (some servers might return this)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceRegistrationResponse{
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Test 200 OK response
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)

	// Generate test keypairs
	signingPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptionPublic := make([]byte, 32)
	_, err = rand.Read(encryptionPublic)
	require.NoError(t, err)

	// Test device registration with 200 OK
	resp, err := client.RegisterDevice("test-token", "Test Device 200", signingPublic, encryptionPublic)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "device-456", resp.Device.DeviceID)
}
