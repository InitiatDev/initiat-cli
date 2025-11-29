package env

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/curve25519"

	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/routes"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func TestShellEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple value",
			input:    "simple",
			expected: `"simple"`,
		},
		{
			name:     "value with quotes",
			input:    `value"with"quotes`,
			expected: `"value\"with\"quotes"`,
		},
		{
			name:     "value with backslash",
			input:    `value\with\backslash`,
			expected: `"value\\with\\backslash"`,
		},
		{
			name:     "value with dollar sign",
			input:    "value$with$dollar",
			expected: `"value\$with\$dollar"`,
		},
		{
			name:     "value with backtick",
			input:    "value`with`backtick",
			expected: "\"value\\`with\\`backtick\"",
		},
		{
			name:     "value with newline",
			input:    "value\nwith\nnewline",
			expected: `"value\nwith\nnewline"`,
		},
		{
			name:     "value with carriage return",
			input:    "value\rwith\rcarriage",
			expected: `"value\rwith\rcarriage"`,
		},
		{
			name:     "value with multiple special chars",
			input:    `value"with\$all\nspecial\chars`,
			expected: `"value\"with\\\$all\\nspecial\\chars"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "value with spaces",
			input:    "value with spaces",
			expected: `"value with spaces"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellEscape(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadEnvironmentSecrets_NoActiveEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := CreateInitiatDir()
	require.NoError(t, err)

	output, err := LoadEnvironmentSecrets("test-org", "test-project")
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "no active environment set")
}

func TestLoadEnvironmentSecrets_NoSecrets(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/environments/dev") {
			env := types.Environment{
				Slug:         "dev",
				Name:         "Development",
				Secrets:      []types.Secret{},
				SecretsCount: 0,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetEnvironmentResponse{Environment: env}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	err := CreateInitiatDir()
	require.NoError(t, err)

	err = CreateEnvironmentDir("dev")
	require.NoError(t, err)

	err = SetActiveEnvironment("dev")
	require.NoError(t, err)

	output, err := LoadEnvironmentSecrets("test-org", "test-project")
	require.NoError(t, err)
	assert.Contains(t, output, "export INITIAT_ENV=")
	assert.Contains(t, output, "dev")
	assert.NotContains(t, output, "export API_KEY=")
}

func TestLoadEnvironmentSecrets_WithSecrets(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	projectKey := make([]byte, 32)
	rand.Read(projectKey)

	encryptedValue1, nonce1, err := crypto.EncryptSecretValue("secret-value-1", projectKey)
	require.NoError(t, err)

	encryptedValue2, nonce2, err := crypto.EncryptSecretValue("secret-value-2", projectKey)
	require.NoError(t, err)

	setupTestEnvironment(t, "")

	store := storage.New()
	devicePrivateKey, err := store.GetEncryptionPrivateKey()
	require.NoError(t, err)

	devicePublicKey, err := curve25519.X25519(devicePrivateKey, curve25519.Basepoint)
	require.NoError(t, err)

	wrappedKey, err := crypto.WrapProjectKey(projectKey, devicePublicKey)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/environments/dev") {
			env := types.Environment{
				Slug: "dev",
				Name: "Development",
				Secrets: []types.Secret{
					{Key: "API_KEY", ID: 1},
					{Key: "DB_PASSWORD", ID: 2},
				},
				SecretsCount: 2,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetEnvironmentResponse{Environment: env}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.Contains(r.URL.Path, "/secrets/API_KEY") {
			secret := types.SecretWithValue{
				Secret: types.Secret{
					Key: "API_KEY",
					ID:  1,
				},
				EncryptedValue: crypto.Encode(encryptedValue1),
				Nonce:          crypto.Encode(nonce1),
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetSecretResponse{Secret: secret}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.Contains(r.URL.Path, "/secrets/DB_PASSWORD") {
			secret := types.SecretWithValue{
				Secret: types.Secret{
					Key: "DB_PASSWORD",
					ID:  2,
				},
				EncryptedValue: crypto.Encode(encryptedValue2),
				Nonce:          crypto.Encode(nonce2),
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetSecretResponse{Secret: secret}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.Contains(r.URL.Path, routes.Project.GetProjectKey("test-org", "test-project")) {
			keyResp := types.GetProjectKeyResponse{
				WrappedProjectKey: wrappedKey,
				KeyVersion:        1,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(keyResp),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	if err := config.Set("api.base_url", server.URL); err != nil {
		t.Fatalf("Failed to set API URL: %v", err)
	}

	err = CreateInitiatDir()
	require.NoError(t, err)

	err = CreateEnvironmentDir("dev")
	require.NoError(t, err)

	err = SetActiveEnvironment("dev")
	require.NoError(t, err)

	output, err := LoadEnvironmentSecrets("test-org", "test-project")
	require.NoError(t, err)

	assert.Contains(t, output, "export INITIAT_ENV=")
	assert.Contains(t, output, "dev")
	assert.Contains(t, output, "export API_KEY=")
	assert.Contains(t, output, "export DB_PASSWORD=")
	assert.Contains(t, output, "secret-value-1")
	assert.Contains(t, output, "secret-value-2")

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 3, len(lines))
}

func TestLoadEnvironmentSecrets_WithSpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	projectKey := make([]byte, 32)
	rand.Read(projectKey)

	specialValue := `value"with\$special\nchars`
	encryptedValue, nonce, err := crypto.EncryptSecretValue(specialValue, projectKey)
	require.NoError(t, err)

	setupTestEnvironment(t, "")

	store := storage.New()
	devicePrivateKey, err := store.GetEncryptionPrivateKey()
	require.NoError(t, err)

	devicePublicKey, err := curve25519.X25519(devicePrivateKey, curve25519.Basepoint)
	require.NoError(t, err)

	wrappedKey, err := crypto.WrapProjectKey(projectKey, devicePublicKey)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/environments/dev") {
			env := types.Environment{
				Slug: "dev",
				Name: "Development",
				Secrets: []types.Secret{
					{Key: "SPECIAL_SECRET", ID: 1},
				},
				SecretsCount: 1,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetEnvironmentResponse{Environment: env}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.Contains(r.URL.Path, "/secrets/SPECIAL_SECRET") {
			secret := types.SecretWithValue{
				Secret: types.Secret{
					Key: "SPECIAL_SECRET",
					ID:  1,
				},
				EncryptedValue: crypto.Encode(encryptedValue),
				Nonce:          crypto.Encode(nonce),
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetSecretResponse{Secret: secret}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.Contains(r.URL.Path, routes.Project.GetProjectKey("test-org", "test-project")) {
			keyResp := types.GetProjectKeyResponse{
				WrappedProjectKey: wrappedKey,
				KeyVersion:        1,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(keyResp),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	if err := config.Set("api.base_url", server.URL); err != nil {
		t.Fatalf("Failed to set API URL: %v", err)
	}

	err = CreateInitiatDir()
	require.NoError(t, err)

	err = CreateEnvironmentDir("dev")
	require.NoError(t, err)

	err = SetActiveEnvironment("dev")
	require.NoError(t, err)

	output, err := LoadEnvironmentSecrets("test-org", "test-project")
	require.NoError(t, err)

	assert.Contains(t, output, "export SPECIAL_SECRET=")
	assert.Contains(t, output, `\"`)
	assert.Contains(t, output, `\\$`)
	assert.Contains(t, output, `\\n`)
}

func TestLoadEnvironmentSecrets_EnvironmentNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	setupTestEnvironment(t, server.URL)

	err := CreateInitiatDir()
	require.NoError(t, err)

	err = CreateEnvironmentDir("dev")
	require.NoError(t, err)

	err = SetActiveEnvironment("dev")
	require.NoError(t, err)

	output, err := LoadEnvironmentSecrets("test-org", "test-project")
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "failed to get environment")
}

func TestLoadEnvironmentSecrets_SecretNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	projectKey := make([]byte, 32)
	rand.Read(projectKey)

	setupTestEnvironment(t, "")

	store := storage.New()
	devicePrivateKey, err := store.GetEncryptionPrivateKey()
	require.NoError(t, err)

	devicePublicKey, err := curve25519.X25519(devicePrivateKey, curve25519.Basepoint)
	require.NoError(t, err)

	wrappedKey, err := crypto.WrapProjectKey(projectKey, devicePublicKey)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/environments/dev") {
			env := types.Environment{
				Slug: "dev",
				Name: "Development",
				Secrets: []types.Secret{
					{Key: "API_KEY", ID: 1},
				},
				SecretsCount: 1,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(types.GetEnvironmentResponse{Environment: env}),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.Contains(r.URL.Path, routes.Project.GetProjectKey("test-org", "test-project")) {
			keyResp := types.GetProjectKeyResponse{
				WrappedProjectKey: wrappedKey,
				KeyVersion:        1,
			}
			response := types.APIResponse{
				Success: true,
				Data:    mustMarshal(keyResp),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	if err := config.Set("api.base_url", server.URL); err != nil {
		t.Fatalf("Failed to set API URL: %v", err)
	}

	err = CreateInitiatDir()
	require.NoError(t, err)

	err = CreateEnvironmentDir("dev")
	require.NoError(t, err)

	err = SetActiveEnvironment("dev")
	require.NoError(t, err)

	output, err := LoadEnvironmentSecrets("test-org", "test-project")
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "failed to get secret")
}

func setupTestEnvironment(t *testing.T, serverURL string) {
	viper.Reset()

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

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
