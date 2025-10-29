package actions

import (
	"testing"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

func TestAssertHTTPAction_Validate(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectStatus int
		retries      *types.Retries
		wantErr      bool
	}{
		{
			name:         "valid HTTP URL",
			url:          "http://example.com",
			expectStatus: 200,
			wantErr:      false,
		},
		{
			name:         "valid HTTPS URL",
			url:          "https://example.com",
			expectStatus: 200,
			wantErr:      false,
		},
		{
			name:         "valid URL with custom status",
			url:          "https://api.example.com/health",
			expectStatus: 201,
			wantErr:      false,
		},
		{
			name:         "valid URL with retries",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			retries:      &types.Retries{Attempts: 3, Backoff: 1},
			wantErr:      false,
		},
		{
			name:         "empty URL",
			url:          "",
			expectStatus: 200,
			wantErr:      true,
		},
		{
			name:         "invalid URL format",
			url:          "not-a-url",
			expectStatus: 200,
			wantErr:      true,
		},
		{
			name:         "invalid status code too low",
			url:          "https://example.com",
			expectStatus: 99,
			wantErr:      true,
		},
		{
			name:         "invalid status code too high",
			url:          "https://example.com",
			expectStatus: 600,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewAssertHTTPAction(tt.url, tt.expectStatus, tt.retries)
			err := action.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertHTTPAction_Render(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectStatus int
		retries      *types.Retries
		os           string
		wantErr      bool
	}{
		{
			name:         "HTTP check on macOS with curl",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			os:           OSMacOS,
			wantErr:      false,
		},
		{
			name:         "HTTP check on Linux with curl",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			os:           OSLinux,
			wantErr:      false,
		},
		{
			name:         "HTTP check on Windows with PowerShell",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			os:           OSWindows,
			wantErr:      true, // PowerShell not available on macOS test environment
		},
		{
			name:         "HTTP check with retries",
			url:          "https://api.example.com/health",
			expectStatus: 200,
			retries:      &types.Retries{Attempts: 3, Backoff: 1},
			os:           OSMacOS,
			wantErr:      false,
		},
		{
			name:         "empty URL",
			url:          "",
			expectStatus: 200,
			os:           OSMacOS,
			wantErr:      true,
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(commands) == 0 {
				t.Error("Expected commands but got none")
			}
		})
	}
}

func TestAssertHTTPAction_StrategyBasedCommands(t *testing.T) {
	action := NewAssertHTTPAction("https://api.example.com/health", 200, nil)
	ctx := &ActionContext{
		OS:         OSMacOS,
		Arch:       "x86_64",
		Env:        map[string]string{},
		WorkingDir: "/tmp",
		Timeout:    30 * time.Second,
	}

	commands, err := action.getHTTPCommands(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Should have one HTTP check command
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}
}
