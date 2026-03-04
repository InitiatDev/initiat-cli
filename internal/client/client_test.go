package client

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/routes"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

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
	client := NewWithBaseURL("http://example.invalid")
	client.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("simulated network error")
	})

	resp, err := client.Login("test@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to make request")
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
	client := NewWithBaseURL("http://example.invalid")
	client.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("simulated network error")
	})

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

func TestGetProjectBySlug_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "project not found"})
	}))
	defer server.Close()

	setupTestEnvironmentForSignedRequests(t, server.URL)
	client := NewWithBaseURL(server.URL)

	project, err := client.GetProjectBySlug("test-org", "non-existent")
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Equal(t, ErrProjectNotFound, err)
}

func TestCreateProject_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-org", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req types.CreateProjectRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "My Project", req.Name)
		assert.Equal(t, "my-project", req.Slug)
		assert.Equal(t, "Test description", req.Description)

		projectData := types.Project{
			ID:             1,
			Name:           "My Project",
			Slug:           "my-project",
			CompositeSlug:  "test-org/my-project",
			Description:    "Test description",
			KeyInitialized: false,
			KeyVersion:     0,
			Role:           "Owner",
			Organization: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
			}{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Project created successfully",
			"data": map[string]interface{}{
				"project": projectData,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironmentForSignedRequests(t, server.URL)
	client := NewWithBaseURL(server.URL)

	project, err := client.CreateProject("test-org", "My Project", "my-project", "Test description")
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "My Project", project.Name)
	assert.Equal(t, "my-project", project.Slug)
	assert.Equal(t, "test-org/my-project", project.CompositeSlug)
	assert.Equal(t, "Test description", project.Description)
	assert.False(t, project.KeyInitialized)
}

func TestCreateProject_EmptyDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateProjectRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "", req.Description)

		projectData := types.Project{
			ID:             1,
			Name:           "My Project",
			Slug:           "my-project",
			CompositeSlug:  "test-org/my-project",
			KeyInitialized: false,
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Project created successfully",
			"data": map[string]interface{}{
				"project": projectData,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironmentForSignedRequests(t, server.URL)
	client := NewWithBaseURL(server.URL)

	project, err := client.CreateProject("test-org", "My Project", "my-project", "")
	assert.NoError(t, err)
	assert.NotNil(t, project)
}

func TestCreateProject_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": false,
			"message": "Project slug already exists",
			"errors":  []string{"Project slug already exists"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironmentForSignedRequests(t, server.URL)
	client := NewWithBaseURL(server.URL)

	project, err := client.CreateProject("test-org", "My Project", "my-project", "")
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "Project slug already exists")
}

func setupTestEnvironmentForSignedRequests(t *testing.T, serverURL string) {
	viper.Reset()

	if err := config.InitConfig(); err != nil {
		t.Fatalf("Failed to init config: %v", err)
	}

	if err := config.Set("api.base_url", serverURL); err != nil {
		t.Fatalf("Failed to set API URL: %v", err)
	}

	serviceName := "initiat-cli-test-" + t.Name()
	if err := config.Set("service_name", serviceName); err != nil {
		t.Fatalf("Failed to set service name: %v", err)
	}

	memKr := storage.NewMemKeyring()
	storage.TestKeyring = memKr
	store := storage.NewWithKeyring(serviceName, memKr)

	signingPublic, signingPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate signing keypair: %v", err)
	}

	if err := store.StoreSigningPrivateKey(signingPrivate); err != nil {
		t.Fatalf("Failed to store signing private key: %v", err)
	}

	encryptionPrivate := make([]byte, 32)
	rand.Read(encryptionPrivate)
	if err := store.StoreEncryptionPrivateKey(encryptionPrivate); err != nil {
		t.Fatalf("Failed to store encryption private key: %v", err)
	}

	if err := store.StoreDeviceID("test-device-123"); err != nil {
		t.Fatalf("Failed to store device ID: %v", err)
	}

	t.Cleanup(func() {
		storage.TestKeyring = nil
		store.DeleteSigningPrivateKey()
		store.DeleteEncryptionPrivateKey()
		store.DeleteDeviceID()
		store.DeleteToken()
	})

	_ = signingPublic
}
