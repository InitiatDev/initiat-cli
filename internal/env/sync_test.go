package env

import (
	"os"
	"testing"
	"time"
)

func TestGetEnvironmentSecrets(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	content := "API_KEY=secret123\nDB_URL=postgres://localhost:5432/test\n# Comment line\n\nEMPTY_VALUE="
	err = WriteSecrets("dev", content)
	if err != nil {
		t.Fatalf("WriteSecrets failed: %v", err)
	}

	secrets, err := GetEnvironmentSecrets("dev")
	if err != nil {
		t.Fatalf("GetEnvironmentSecrets failed: %v", err)
	}

	if len(secrets) != 3 {
		t.Errorf("Expected 3 secrets, got %d", len(secrets))
	}

	expectedKeys := []string{"API_KEY", "DB_URL", "EMPTY_VALUE"}
	for i, secret := range secrets {
		if secret.Key != expectedKeys[i] {
			t.Errorf("Expected key %s, got %s", expectedKeys[i], secret.Key)
		}
	}
}

func TestGetEnvironmentSecretsEmpty(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	secrets, err := GetEnvironmentSecrets("dev")
	if err != nil {
		t.Fatalf("GetEnvironmentSecrets failed: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("Expected 0 secrets, got %d", len(secrets))
	}
}

func TestGetLastSyncTime(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	err := CreateInitiatDir()
	if err != nil {
		t.Fatalf("CreateInitiatDir failed: %v", err)
	}

	before := time.Now()

	err = WriteSecrets("dev", "API_KEY=secret123")
	if err != nil {
		t.Fatalf("WriteSecrets failed: %v", err)
	}

	after := time.Now()

	syncTime, err := GetLastSyncTime("dev")
	if err != nil {
		t.Fatalf("GetLastSyncTime failed: %v", err)
	}

	if syncTime.Before(before) || syncTime.After(after) {
		t.Errorf("Sync time %v not between %v and %v", syncTime, before, after)
	}
}

func TestGetLastSyncTimeNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	_, err := GetLastSyncTime("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent environment")
	}
}
