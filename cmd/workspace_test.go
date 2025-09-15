package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DylanBlakemore/initflow-cli/internal/client"
	"github.com/DylanBlakemore/initflow-cli/internal/config"
	"github.com/DylanBlakemore/initflow-cli/internal/encoding"
	"github.com/DylanBlakemore/initflow-cli/internal/storage"
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
		response := client.ListWorkspacesResponse{
			Workspaces: []client.Workspace{
				{
					ID:             1,
					Name:           "My Project",
					Slug:           "my-project",
					Description:    "Test project",
					KeyInitialized: false,
					KeyVersion:     0,
					Role:           "Owner",
					Account: struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
						Slug string `json:"slug"`
					}{
						ID:   123,
						Name: "Test Account",
						Slug: "test-account",
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
					Account: struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
						Slug string `json:"slug"`
					}{
						ID:   456,
						Name: "Team Account",
						Slug: "team-account",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	err := runWorkspaceList(workspaceListCmd, []string{})
	if err != nil {
		t.Fatalf("runWorkspaceList failed: %v", err)
	}
}

func TestWorkspaceInitKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/workspaces":
			response := client.ListWorkspacesResponse{
				Workspaces: []client.Workspace{
					{
						ID:             1,
						Name:           "My Project",
						Slug:           "my-project",
						Description:    "Test project",
						KeyInitialized: false,
						KeyVersion:     0,
						Role:           "Owner",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case "/api/v1/workspaces/1/initialize":
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}

			var req client.InitializeWorkspaceKeyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			if req.WrappedWorkspaceKey == "" {
				t.Error("Expected wrapped workspace key")
			}

			if _, err := encoding.Decode(req.WrappedWorkspaceKey); err != nil {
				t.Errorf("Invalid encoded wrapped key: %v", err)
			}
			response := client.InitializeWorkspaceKeyResponse{
				Success: true,
				Message: "Workspace key initialized successfully",
				Workspace: client.Workspace{
					ID:             1,
					Name:           "My Project",
					Slug:           "my-project",
					KeyInitialized: true,
					KeyVersion:     1,
				},
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

	err := runWorkspaceInit(workspaceInitCmd, []string{"my-project"})
	if err != nil {
		t.Fatalf("runWorkspaceInit failed: %v", err)
	}

	store := storage.New()
	if !store.HasWorkspaceKey("my-project") {
		t.Error("Expected workspace key to be stored locally")
	}
}

func TestWorkspaceInitKeyAlreadyInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ListWorkspacesResponse{
			Workspaces: []client.Workspace{
				{
					ID:             1,
					Name:           "My Project",
					Slug:           "my-project",
					Description:    "Test project",
					KeyInitialized: true,
					KeyVersion:     1,
					Role:           "Owner",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	err := runWorkspaceInit(workspaceInitCmd, []string{"my-project"})
	if err == nil {
		t.Error("Expected error for already initialized workspace")
	}
	if err.Error() != "ℹ️ Workspace key already initialized" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestWorkspaceInitKeyNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ListWorkspacesResponse{
			Workspaces: []client.Workspace{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	err := runWorkspaceInit(workspaceInitCmd, []string{"non-existent"})
	if err == nil {
		t.Error("Expected error for non-existent workspace")
	}
	if err.Error() != "❌ Failed to get workspace info: workspace 'non-existent' not found" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestWrapWorkspaceKey(t *testing.T) {
	store := storage.New()

	encryptionPrivate := make([]byte, 32)
	rand.Read(encryptionPrivate)
	store.StoreEncryptionPrivateKey(encryptionPrivate)
	defer store.DeleteEncryptionPrivateKey()

	workspaceKey := make([]byte, 32)
	rand.Read(workspaceKey)

	wrappedKey, err := wrapWorkspaceKey(workspaceKey, store)
	if err != nil {
		t.Fatalf("wrapWorkspaceKey failed: %v", err)
	}

	if len(wrappedKey) < 32+12+32 {
		t.Errorf("Wrapped key too short: %d bytes", len(wrappedKey))
	}

	if len(wrappedKey) < 44 {
		t.Fatal("Wrapped key too short to extract components")
	}

	ephemeralPublic := wrappedKey[:32]
	nonce := wrappedKey[32:44]
	ciphertext := wrappedKey[44:]

	if len(ephemeralPublic) != 32 {
		t.Errorf("Expected 32-byte ephemeral public key, got %d", len(ephemeralPublic))
	}
	if len(nonce) != 12 {
		t.Errorf("Expected 12-byte nonce, got %d", len(nonce))
	}
	if len(ciphertext) < 32 {
		t.Errorf("Expected at least 32-byte ciphertext, got %d", len(ciphertext))
	}
}

func setupTestEnvironment(t *testing.T, serverURL string) {
	if err := config.InitConfig(); err != nil {
		t.Fatalf("Failed to init config: %v", err)
	}

	if err := config.Set("api_base_url", serverURL); err != nil {
		t.Fatalf("Failed to set API URL: %v", err)
	}

	if err := config.Set("service_name", "initflow-cli-test-"+t.Name()); err != nil {
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
		store.DeleteWorkspaceKey("my-project")
		store.DeleteWorkspaceKey("team-secrets")
		store.DeleteWorkspaceKey("personal-vault")
		store.DeleteWorkspaceKey("non-existent")
	})

	oldStdout := os.Stdout
	os.Stdout = nil
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	_ = signingPublic
}
