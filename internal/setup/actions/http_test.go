package actions

import (
	"strings"
	"testing"
	"time"
)

func TestAssertHTTPAction_Validate(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectStatus int
		retries      *Retries
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid HTTP URL",
			url:          "http://example.com",
			expectStatus: 200,
			expectError:  false,
		},
		{
			name:         "valid HTTPS URL",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			expectError:  false,
		},
		{
			name:         "valid URL with custom status",
			url:          "https://api.example.com/status",
			expectStatus: 201,
			expectError:  false,
		},
		{
			name:         "valid URL with retries",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			retries: &Retries{
				Attempts: 3,
				Backoff:  "2s",
				MaxDelay: "10s",
			},
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
			errorMsg:    "URL cannot be empty",
		},
		{
			name:        "invalid URL format",
			url:         "ftp://example.com",
			expectError: true,
			errorMsg:    "URL must start with http:// or https://",
		},
		{
			name:         "invalid status code too low",
			url:          "https://example.com",
			expectStatus: 99,
			expectError:  true,
			errorMsg:     "expected status code must be between 100 and 599",
		},
		{
			name:         "invalid status code too high",
			url:          "https://example.com",
			expectStatus: 600,
			expectError:  true,
			errorMsg:     "expected status code must be between 100 and 599",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewAssertHTTPAction(tt.url, tt.expectStatus, tt.retries)
			err := action.Validate()

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAssertHTTPAction_Render(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectStatus int
		retries      *Retries
		os           string
		expectError  bool
	}{
		{
			name:         "HTTP check on macOS with curl",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			os:           OSMacOS,
			expectError:  false,
		},
		{
			name:         "HTTP check on Linux with curl",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			os:           OSLinux,
			expectError:  false,
		},
		{
			name:         "HTTP check on Windows with PowerShell",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			os:           OSWindows,
			expectError:  false,
		},
		{
			name:         "HTTP check with retries",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			retries: &Retries{
				Attempts: 3,
				Backoff:  "2s",
				MaxDelay: "10s",
			},
			os:          OSMacOS,
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewAssertHTTPAction(tt.url, tt.expectStatus, tt.retries)
			ctx := &ActionContext{
				OS:         tt.os,
				Arch:       "x86_64",
				Env:        map[string]string{},
				WorkingDir: "/tmp",
				Timeout:    30 * time.Second,
			}

			commands, err := action.Render(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(commands) == 0 {
					t.Error("Expected commands but got none")
				}
			}
		})
	}
}

func TestAssertHTTPAction_GetCurlArgs(t *testing.T) {
	action := &AssertHTTPAction{
		url:          "https://api.example.com/health",
		expectStatus: 200,
	}

	args := action.getCurlArgs()
	if len(args) == 0 {
		t.Fatal("Expected curl arguments but got none")
	}

	// Check for essential curl arguments
	expectedArgs := []string{"-s", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "30"}
	for _, expected := range expectedArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected curl argument '%s' not found", expected)
		}
	}

	// Check that URL is included
	urlFound := false
	for _, arg := range args {
		if arg == action.url {
			urlFound = true
			break
		}
	}
	if !urlFound {
		t.Error("Expected URL to be included in curl arguments")
	}
}

func TestAssertHTTPAction_GetCurlArgsWithRetries(t *testing.T) {
	retries := &Retries{
		Attempts: 3,
		Backoff:  "2s",
		MaxDelay: "10s",
	}
	action := &AssertHTTPAction{
		url:          "https://api.example.com/health",
		expectStatus: 200,
		retries:      retries,
	}

	args := action.getCurlArgs()
	if len(args) == 0 {
		t.Fatal("Expected curl arguments but got none")
	}

	// Check for retry arguments
	expectedRetryArgs := []string{"--retry", "2", "--retry-delay", "1", "--retry-max-time", "60"}
	for _, expected := range expectedRetryArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected curl retry argument '%s' not found", expected)
		}
	}
}

func TestAssertHTTPAction_GetWgetArgs(t *testing.T) {
	action := &AssertHTTPAction{
		url:          "https://api.example.com/health",
		expectStatus: 200,
	}

	args := action.getWgetArgs()
	if len(args) == 0 {
		t.Fatal("Expected wget arguments but got none")
	}

	// Check for essential wget arguments
	expectedArgs := []string{"--spider", "--quiet", "--timeout=30", "--tries=1"}
	for _, expected := range expectedArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected wget argument '%s' not found", expected)
		}
	}

	// Check that URL is included
	urlFound := false
	for _, arg := range args {
		if arg == action.url {
			urlFound = true
			break
		}
	}
	if !urlFound {
		t.Error("Expected URL to be included in wget arguments")
	}
}

func TestAssertHTTPAction_GetWgetArgsWithRetries(t *testing.T) {
	retries := &Retries{
		Attempts: 3,
		Backoff:  "2s",
		MaxDelay: "10s",
	}
	action := &AssertHTTPAction{
		url:          "https://api.example.com/health",
		expectStatus: 200,
		retries:      retries,
	}

	args := action.getWgetArgs()
	if len(args) == 0 {
		t.Fatal("Expected wget arguments but got none")
	}

	// Check for retry arguments
	expectedRetryArgs := []string{"--tries", "3"}
	for _, expected := range expectedRetryArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected wget retry argument '%s' not found", expected)
		}
	}
}

func TestAssertHTTPAction_GetPowerShellArgs(t *testing.T) {
	action := &AssertHTTPAction{
		url:          "https://api.example.com/health",
		expectStatus: 200,
	}

	args := action.getPowerShellArgs()
	if len(args) == 0 {
		t.Fatal("Expected PowerShell arguments but got none")
	}

	// Check for PowerShell command structure
	if args[0] != "-Command" {
		t.Error("Expected first argument to be '-Command'")
	}

	// Check that the script contains the URL
	script := args[1]
	if !strings.Contains(script, action.url) {
		t.Error("Expected PowerShell script to contain the URL")
	}

	// Check that the script contains Invoke-WebRequest
	if !strings.Contains(script, "Invoke-WebRequest") {
		t.Error("Expected PowerShell script to contain Invoke-WebRequest")
	}
}

func TestAssertHTTPAction_CreateRetryScript(t *testing.T) {
	retries := &Retries{
		Attempts: 3,
		Backoff:  "2s",
		MaxDelay: "10s",
	}
	action := &AssertHTTPAction{
		url:          "https://api.example.com/health",
		expectStatus: 200,
		retries:      retries,
	}

	script := action.createRetryScript("curl", []string{"-s", "-o", "/dev/null", "-w", "%{http_code}", "https://api.example.com/health"})
	if script == "" {
		t.Fatal("Expected retry script but got empty string")
	}

	// Check that script contains expected elements
	expectedElements := []string{
		"#!/bin/bash",
		"URL=\"https://api.example.com/health\"",
		"EXPECTED_STATUS=200",
		"ATTEMPTS=3",
		"BACKOFF=2",
		"MAX_DELAY=30",
		"curl -s -o /dev/null -w",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(script, expected) {
			t.Errorf("Expected retry script to contain '%s'", expected)
		}
	}
}

func TestAssertHTTPAction_IsCommandAvailable(t *testing.T) {
	// Test with a command that should exist
	action := &AssertHTTPAction{}

	// This test is platform-dependent, so we'll just test the function exists
	// and doesn't panic
	_ = action.isCommandAvailable("ls")
	_ = action.isCommandAvailable("nonexistentcommand")
}
