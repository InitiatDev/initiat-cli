package project

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func TestSetupProject_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/projects/test-org/my-project":
			project := types.Project{
				ID:             1,
				Name:           "My Project",
				Slug:           "my-project",
				CompositeSlug:  "test-org/my-project",
				Description:    "Test project",
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
				"message": "Project retrieved successfully",
				"data":    project,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case "/api/v1/projects/test-org/my-project/initialize":
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}

			var req types.InitializeProjectKeyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			if req.WrappedProjectKey == "" {
				t.Error("Expected wrapped project key")
			}

			response := map[string]interface{}{
				"success": true,
				"message": "Project key initialized successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		default:
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	details := SetupDetails{
		OrgSlug:     "test-org",
		ProjectSlug: "my-project",
	}

	result, err := SetupProject(details)
	if err != nil {
		t.Fatalf("SetupProject failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if !result.KeyInitialized {
		t.Error("Expected key to be initialized")
	}

	if result.ProjectCreated {
		t.Error("Expected project not to be created (it already exists)")
	}

	if !strings.Contains(result.Message, "Project key initialized successfully") {
		t.Errorf("Expected success message, got: %s", result.Message)
	}
}

func TestSetupProject_ProjectNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "project not found"})
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	details := SetupDetails{
		OrgSlug:     "test-org",
		ProjectSlug: "non-existent",
	}

	result, err := SetupProject(details)
	if err != nil {
		t.Fatalf("SetupProject failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.KeyInitialized {
		t.Error("Expected key not to be initialized")
	}

	if result.ProjectCreated {
		t.Error("Expected project not to be created")
	}

	if !strings.Contains(result.Message, "doesn't exist remotely") {
		t.Errorf("Expected project not found message, got: %s", result.Message)
	}
}

func TestSetupProject_DeviceNotRegistered(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	if err := config.InitConfig(); err != nil {
		t.Fatalf("Failed to init config: %v", err)
	}

	details := SetupDetails{
		OrgSlug:     "test-org",
		ProjectSlug: "my-project",
	}

	_, err := SetupProject(details)
	if err == nil {
		t.Error("Expected error for unregistered device")
	}

	if !strings.Contains(err.Error(), "device not registered") {
		t.Errorf("Expected device registration error, got: %v", err)
	}
}

func TestCreateInitiatFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	err := CreateInitiatFile("test-org", "test-project")
	if err != nil {
		t.Fatalf("Failed to create .initiat file: %v", err)
	}

	initiatPath := filepath.Join(tempDir, ".initiat")
	if _, err := os.Stat(initiatPath); os.IsNotExist(err) {
		t.Error("Expected .initiat file to be created")
	}

	content, err := os.ReadFile(initiatPath)
	if err != nil {
		t.Fatalf("Failed to read .initiat file: %v", err)
	}

	expectedContent := "org: test-org\nproject: test-project\n"
	if string(content) != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, string(content))
	}
}

func TestCheckInitiatFileExists(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	exists, err := CheckInitiatFileExists()
	if err != nil {
		t.Fatalf("Failed to check .initiat file: %v", err)
	}
	if exists {
		t.Error("Expected .initiat file not to exist")
	}

	initiatPath := filepath.Join(tempDir, ".initiat")
	err = os.WriteFile(initiatPath, []byte("org: test-org\nproject: test-project\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .initiat file: %v", err)
	}

	exists, err = CheckInitiatFileExists()
	if err != nil {
		t.Fatalf("Failed to check .initiat file: %v", err)
	}
	if !exists {
		t.Error("Expected .initiat file to exist")
	}
}

func TestGetDefaultProjectName(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	name, err := GetDefaultProjectName()
	if err != nil {
		t.Fatalf("Failed to get default project name: %v", err)
	}

	expectedName := filepath.Base(tempDir)
	if name != expectedName {
		t.Errorf("Expected default project name %q, got %q", expectedName, name)
	}
}

func TestSetupProject_KeyAlreadyInitialized(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/projects/test-org/my-project" {
			project := types.Project{
				ID:             1,
				Name:           "My Project",
				Slug:           "my-project",
				CompositeSlug:  "test-org/my-project",
				Description:    "Test project",
				KeyInitialized: true,
				KeyVersion:     1,
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
				"message": "Project retrieved successfully",
				"data":    project,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	details := SetupDetails{
		OrgSlug:     "test-org",
		ProjectSlug: "my-project",
	}

	result, err := SetupProject(details)
	if err != nil {
		t.Fatalf("SetupProject failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if !result.KeyInitialized {
		t.Error("Expected key to be initialized")
	}

	if result.ProjectCreated {
		t.Error("Expected project not to be created")
	}

	if !strings.Contains(result.Message, "Project key already initialized") {
		t.Errorf("Expected already initialized message, got: %s", result.Message)
	}
}

func setupTestEnvironment(t *testing.T, serverURL string) {
	if err := config.InitConfig(); err != nil {
		t.Fatalf("Failed to init config: %v", err)
	}

	if err := config.Set("api.base_url", serverURL); err != nil {
		t.Fatalf("Failed to set API URL: %v", err)
	}

	if err := config.Set("service_name", "initiat-cli-test-"+t.Name()); err != nil {
		t.Fatalf("Failed to set service name: %v", err)
	}

	store := storage.New()

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
		store.DeleteSigningPrivateKey()
		store.DeleteEncryptionPrivateKey()
		store.DeleteDeviceID()
		store.DeleteToken()
	})

	_ = signingPublic
}
