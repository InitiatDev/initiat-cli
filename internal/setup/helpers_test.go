package setup

import (
	"os"
	"testing"
)

func TestCollectSecretNames(t *testing.T) {
	tests := []struct {
		name     string
		config   *SetupConfig
		expected []string
	}{
		{
			name: "no secrets",
			config: &SetupConfig{
				Version: 1,
				Setup: []Step{
					{Run: "echo test"},
				},
			},
			expected: []string{},
		},
		{
			name: "single secret",
			config: &SetupConfig{
				Version: 1,
				Setup: []Step{
					{
						Run:     "echo test",
						Secrets: []string{"API_KEY"},
					},
				},
			},
			expected: []string{"API_KEY"},
		},
		{
			name: "multiple secrets in multiple steps",
			config: &SetupConfig{
				Version: 1,
				Setup: []Step{
					{
						Run:     "echo test",
						Secrets: []string{"API_KEY", "DATABASE_URL"},
					},
					{
						Run:     "echo test2",
						Secrets: []string{"DATABASE_URL", "SECRET_TOKEN"},
					},
				},
			},
			expected: []string{"API_KEY", "DATABASE_URL", "SECRET_TOKEN"},
		},
		{
			name: "duplicate secrets",
			config: &SetupConfig{
				Version: 1,
				Setup: []Step{
					{
						Run:     "echo test",
						Secrets: []string{"API_KEY"},
					},
					{
						Run:     "echo test2",
						Secrets: []string{"API_KEY"},
					},
				},
			},
			expected: []string{"API_KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectSecretNames(tt.config)

			if len(result) != len(tt.expected) {
				t.Errorf("collectSecretNames() returned %d secrets, want %d", len(result), len(tt.expected))
				return
			}

			secretMap := make(map[string]bool)
			for _, secret := range result {
				secretMap[secret] = true
			}

			for _, expected := range tt.expected {
				if !secretMap[expected] {
					t.Errorf("collectSecretNames() missing secret: %s", expected)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOfString(s, substr) >= 0)))
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestDetectShell(t *testing.T) {
	originalShell := os.Getenv("SHELL")

	tests := []struct {
		name     string
		setShell string
		want     string
	}{
		{
			name:     "with SHELL env var",
			setShell: "/bin/zsh",
			want:     "/bin/zsh",
		},
		{
			name:     "without SHELL env var",
			setShell: "",
			want:     "/bin/bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setShell != "" {
				os.Setenv("SHELL", tt.setShell)
			} else {
				os.Unsetenv("SHELL")
			}
			defer func() {
				if originalShell != "" {
					os.Setenv("SHELL", originalShell)
				} else {
					os.Unsetenv("SHELL")
				}
			}()

			result := detectShell()

			if result != tt.want {
				t.Errorf("detectShell() = %q, want %q", result, tt.want)
			}
		})
	}
}
