package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/testutil"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func TestProjectList(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/projects" {
			t.Errorf("Expected GET /api/v1/projects, got %s %s", r.Method, r.URL.Path)
		}

		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header")
		}
		if r.Header.Get("X-Signature") == "" {
			t.Error("Expected X-Signature header")
		}
		if r.Header.Get("X-Timestamp") == "" {
			t.Error("Expected X-Timestamp header")
		}
		projectsData := types.ListProjectsResponse{
			Projects: []types.Project{
				{
					ID:             1,
					Name:           "My Project",
					Slug:           "my-project",
					Description:    "Test project",
					KeyInitialized: false,
					KeyVersion:     0,
					Role:           "Owner",
					Organization: struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
						Slug string `json:"slug"`
					}{
						ID:   123,
						Name: "Test Organization",
						Slug: "test-organization",
					},
				},
				{
					ID:             2,
					Name:           "Team Secrets",
					Slug:           "team-secrets",
					Description:    "Team project",
					KeyInitialized: true,
					KeyVersion:     1,
					Role:           "Member",
					Organization: struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
						Slug string `json:"slug"`
					}{
						ID:   456,
						Name: "Team Organization",
						Slug: "team-organization",
					},
				},
			},
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Projects retrieved successfully",
			"data":    projectsData,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	err := runProjectList(projectListCmd, []string{})
	if err != nil {
		t.Fatalf("runProjectList failed: %v", err)
	}

	capture.AssertContains(t, "My Project")
}

func TestProjectInitKey(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

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

			if _, err := crypto.Decode(req.WrappedProjectKey); err != nil {
				t.Errorf("Invalid encoded wrapped key: %v", err)
			}
			projectData := types.Project{
				ID:             1,
				Name:           "My Project",
				Slug:           "my-project",
				KeyInitialized: true,
				KeyVersion:     1,
			}

			response := map[string]interface{}{
				"success": true,
				"message": "Project key initialized successfully",
				"data":    map[string]interface{}{"project": projectData},
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

	projectPath = "test-org/my-project"
	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("runProjectInit failed: %v", err)
	}
}

func TestProjectInitKeyAlreadyInitialized(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/projects/test-org/my-project") {
			project := types.Project{
				ID:             1,
				Name:           "My Project",
				Slug:           "my-project",
				Description:    "Test project",
				KeyInitialized: true,
				KeyVersion:     1,
				Role:           "Owner",
			}

			response := map[string]interface{}{
				"success": true,
				"message": "Project retrieved successfully",
				"data":    project,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	projectPath = "test-org/my-project"
	_ = runProjectInit(projectInitCmd, []string{})
}

func TestProjectInitKeyNotFound_UserDeclines(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	mock, err := testutil.MockStdin("n\n")
	if err != nil {
		t.Fatalf("Failed to mock stdin: %v", err)
	}
	defer mock.Restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "project not found"})
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	projectPath = "test-org/non-existent"
	err = runProjectInit(projectInitCmd, []string{})
	if err == nil {
		t.Error("Expected error when user declines to create project")
		return
	}
	if !strings.Contains(err.Error(), "Project creation cancelled") {
		t.Errorf("Expected 'Project creation cancelled' error, got: %v", err)
	}
	capture.AssertContains(t, "does not exist")
	capture.AssertContains(t, "Would you like to create it?")
}

func TestProjectInitKeyNotFound_UserCreates(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	mock, err := testutil.MockStdin("y\n")
	if err != nil {
		t.Fatalf("Failed to mock stdin: %v", err)
	}
	defer mock.Restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/projects/test-org":
			if r.Method == "POST" {
				var req types.CreateProjectRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request: %v", err)
				}

				projectData := types.Project{
					ID:             1,
					Name:           "New Project",
					Slug:           "new-project",
					CompositeSlug:  "test-org/new-project",
					Description:    "",
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
				return
			}
		case "/api/v1/projects/test-org/new-project":
			if r.Method == "GET" {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"error": "project not found"})
				return
			}
		case "/api/v1/projects/test-org/new-project/initialize":
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

			if _, err := crypto.Decode(req.WrappedProjectKey); err != nil {
				t.Errorf("Invalid encoded wrapped key: %v", err)
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

	projectPath = "test-org/new-project"
	err = runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("runProjectInit failed: %v", err)
	}

	capture.AssertContains(t, "does not exist")
	capture.AssertContains(t, "Would you like to create it?")
	capture.AssertContains(t, "Creating project")
	capture.AssertContains(t, "created successfully")
	capture.AssertContains(t, "initialized successfully")
}

func setupTestEnvironment(t *testing.T, serverURL string) {
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
