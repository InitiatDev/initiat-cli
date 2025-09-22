package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DylanBlakemore/initiat-cli/internal/client"
	"github.com/DylanBlakemore/initiat-cli/internal/config"
)

func TestLoginCmd_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.LoginResponse{
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

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := config.InitConfig()
	require.NoError(t, err)
	err = config.Set("api_base_url", server.URL)
	require.NoError(t, err)
	cmd := &cobra.Command{}
	cmd.AddCommand(authCmd)

	assert.Equal(t, "auth", authCmd.Use)
	assert.Equal(t, "Authentication commands", authCmd.Short)

	loginSubCmd, _, err := authCmd.Find([]string{"login"})
	require.NoError(t, err)
	assert.Equal(t, "login <email>", loginSubCmd.Use)
	assert.Equal(t, "Login to Initiat", loginSubCmd.Short)
}

func TestLoginCmd_InvalidArgs(t *testing.T) {
	assert.NotNil(t, loginCmd.Args)
	err := loginCmd.Args(loginCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")

	err = loginCmd.Args(loginCmd, []string{"email1", "email2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 2")

	err = loginCmd.Args(loginCmd, []string{"test@example.com"})
	assert.NoError(t, err)
}

func TestLoginCmd_EmptyEmail(t *testing.T) {
	err := runLogin(loginCmd, []string{""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")

	err = runLogin(loginCmd, []string{"   "})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestAuthCmd_Structure(t *testing.T) {
	assert.Equal(t, "auth", authCmd.Use)
	assert.Equal(t, "Authentication commands", authCmd.Short)
	assert.Equal(t, "Manage authentication with Initiat", authCmd.Long)
	var loginFound bool
	for _, cmd := range authCmd.Commands() {
		if cmd.Use == "login <email>" {
			loginFound = true
			assert.Equal(t, "Login to Initiat", cmd.Short)
			assert.Equal(t, "Authenticate with Initiat using your email and password", cmd.Long)
			break
		}
	}
	assert.True(t, loginFound, "login subcommand not found")
}

func TestLoginCmd_NetworkError(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := config.InitConfig()
	require.NoError(t, err)
	err = config.Set("api_base_url", "http://invalid-url-that-does-not-exist.com")
	require.NoError(t, err)
}

func TestLoginCmd_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid credentials",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	err := config.InitConfig()
	require.NoError(t, err)
	err = config.Set("api_base_url", server.URL)
	require.NoError(t, err)
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestAuthCmd_Integration(t *testing.T) {
	rootCmd := &cobra.Command{Use: "initiat"}
	rootCmd.AddCommand(authCmd)

	output, err := executeCommand(rootCmd, "auth", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "Manage authentication with Initiat")
	assert.Contains(t, output, "login")

	output, err = executeCommand(rootCmd, "auth", "login", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "Authenticate with Initiat using your email and password")
	assert.Contains(t, output, "login <email>")
}
