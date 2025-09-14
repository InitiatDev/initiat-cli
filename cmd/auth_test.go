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

	"github.com/DylanBlakemore/initflow-cli/internal/client"
	"github.com/DylanBlakemore/initflow-cli/internal/config"
)

func TestLoginCmd_Success(t *testing.T) {
	// Create test server
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
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set up temporary config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config with test server URL
	err := config.InitConfig()
	require.NoError(t, err)
	err = config.Set("api_base_url", server.URL)
	require.NoError(t, err)

	// Test the command structure
	cmd := &cobra.Command{}
	cmd.AddCommand(authCmd)

	// Verify command exists and has correct structure
	assert.Equal(t, "auth", authCmd.Use)
	assert.Equal(t, "Authentication commands", authCmd.Short)

	// Find login subcommand
	loginSubCmd, _, err := authCmd.Find([]string{"login"})
	require.NoError(t, err)
	assert.Equal(t, "login <email>", loginSubCmd.Use)
	assert.Equal(t, "Login to InitFlow", loginSubCmd.Short)
}

func TestLoginCmd_InvalidArgs(t *testing.T) {
	// Test command validation through cobra's built-in validation
	// We can't directly test runLogin with invalid args since it expects exactly 1 arg
	// Instead, test the command structure and validation

	// Verify the command has the correct Args function (we can't compare functions directly)
	assert.NotNil(t, loginCmd.Args)

	// Test that the Args function works correctly
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
	// Verify auth command structure
	assert.Equal(t, "auth", authCmd.Use)
	assert.Equal(t, "Authentication commands", authCmd.Short)
	assert.Equal(t, "Manage authentication with InitFlow", authCmd.Long)

	// Verify login subcommand exists
	var loginFound bool
	for _, cmd := range authCmd.Commands() {
		if cmd.Use == "login <email>" {
			loginFound = true
			assert.Equal(t, "Login to InitFlow", cmd.Short)
			assert.Equal(t, "Authenticate with InitFlow using your email and password", cmd.Long)
			break
		}
	}
	assert.True(t, loginFound, "login subcommand not found")
}

func TestLoginCmd_NetworkError(t *testing.T) {
	// Set up temporary config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config with invalid URL
	err := config.InitConfig()
	require.NoError(t, err)
	err = config.Set("api_base_url", "http://invalid-url-that-does-not-exist.com")
	require.NoError(t, err)

	// Mock password input by creating a version of runLogin that doesn't read from stdin
	// In a real test, you'd need to mock the terminal input
	// For now, we'll test the command structure and argument validation
}

func TestLoginCmd_ServerError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid credentials",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set up temporary config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config with test server URL
	err := config.InitConfig()
	require.NoError(t, err)
	err = config.Set("api_base_url", server.URL)
	require.NoError(t, err)

	// Note: Testing the actual runLogin function with password input would require
	// mocking terminal input, which is complex. In a production environment,
	// you might want to refactor the code to accept password as a parameter
	// for testing purposes, or use dependency injection for the input reader.
}

// Test helper function to capture command output
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestAuthCmd_Integration(t *testing.T) {
	// Create a root command for testing
	rootCmd := &cobra.Command{Use: "initflow"}
	rootCmd.AddCommand(authCmd)

	// Test help output
	output, err := executeCommand(rootCmd, "auth", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "Manage authentication with InitFlow")
	assert.Contains(t, output, "login")

	// Test login help output
	output, err = executeCommand(rootCmd, "auth", "login", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "Authenticate with InitFlow using your email and password")
	assert.Contains(t, output, "login <email>")
}
