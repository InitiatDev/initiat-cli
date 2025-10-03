package cmd

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func TestWorkspaceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/workspaces" {
			t.Errorf("Expected GET /api/v1/workspaces, got %s %s", r.Method, r.URL.Path)
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
		workspacesData := types.ListWorkspacesResponse{
			Workspaces: []types.Workspace{
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
			"message": "Workspaces retrieved successfully",
			"data":    workspacesData,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorkspaceList(workspaceListCmd, []string{})
	if err != nil {
		w.Close()
		os.Stdout = old
		t.Fatalf("runWorkspaceList failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "My Project") {
		t.Error("Expected 'My Project' in output")
	}
}

func TestWorkspaceInitKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/workspaces/test-org/my-project":
			workspace := types.Workspace{
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
				"message": "Workspace retrieved successfully",
				"data":    workspace,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case "/api/v1/workspaces/test-org/my-project/initialize":
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}

			var req types.InitializeWorkspaceKeyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			if req.WrappedWorkspaceKey == "" {
				t.Error("Expected wrapped workspace key")
			}

			if _, err := crypto.Decode(req.WrappedWorkspaceKey); err != nil {
				t.Errorf("Invalid encoded wrapped key: %v", err)
			}
			workspaceData := types.Workspace{
				ID:             1,
				Name:           "My Project",
				Slug:           "my-project",
				KeyInitialized: true,
				KeyVersion:     1,
			}

			response := map[string]interface{}{
				"success": true,
				"message": "Workspace key initialized successfully",
				"data":    map[string]interface{}{"workspace": workspaceData},
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

	workspacePath = "test-org/my-project"
	err := runWorkspaceInit(workspaceInitCmd, []string{})
	if err != nil {
		t.Fatalf("runWorkspaceInit failed: %v", err)
	}
}

func TestWorkspaceInitKeyAlreadyInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/workspaces/test-org/my-project") {
			workspace := types.Workspace{
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
				"message": "Workspace retrieved successfully",
				"data":    workspace,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	workspacePath = "test-org/my-project"
	_ = runWorkspaceInit(workspaceInitCmd, []string{})
}

func TestWorkspaceInitKeyNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "workspace not found"})
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	workspacePath = "test-org/non-existent"
	err := runWorkspaceInit(workspaceInitCmd, []string{})
	if err == nil {
		t.Error("Expected error for non-existent workspace")
		return
	}
	if !strings.Contains(err.Error(), "Failed to get workspace info") {
		t.Errorf("Expected specific error message, got: %v", err)
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

	oldStdout := os.Stdout
	os.Stdout = nil
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	_ = signingPublic
}
