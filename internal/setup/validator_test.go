package setup

import (
	"strings"
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Name:    "Test Setup",
		Env: &Environment{
			Vars:    map[string]string{"NODE_ENV": "development"},
			Secrets: []string{"DATABASE_URL"},
		},
		Setup: []Step{
			{
				Name: "Install deps",
				Run:  "npm install",
			},
			{
				Name:           "Setup database",
				Run:            "mix ecto.setup",
				EnvFromSecrets: []string{"DATABASE_URL"},
			},
		},
	}

	err := Validate(config)
	if err != nil {
		t.Errorf("Expected no validation errors, got: %v", err)
	}
}

func TestValidate_InvalidVersion(t *testing.T) {
	config := &SetupConfig{
		Version: 2,
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for invalid version")
	}

	if !strings.Contains(err.Error(), "version") {
		t.Errorf("Expected version error, got: %v", err)
	}
}

func TestValidate_StepWithoutAction(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "Empty step",
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for step without action")
	}

	if !strings.Contains(err.Error(), "exactly one action") {
		t.Errorf("Expected action error, got: %v", err)
	}
}

func TestValidate_StepWithMultipleActions(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name:  "Multiple actions",
				Run:   "echo hello",
				Print: "This should fail",
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for step with multiple actions")
	}

	if !strings.Contains(err.Error(), "multiple actions") {
		t.Errorf("Expected multiple actions error, got: %v", err)
	}
}

func TestValidate_InvalidTimeout(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name:    "Invalid timeout",
				Run:     "echo hello",
				Timeout: "invalid",
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for invalid timeout")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestValidate_InvalidRetries(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "Invalid retries",
				Run:  "echo hello",
				Retries: &Retries{
					Attempts: 0,
					Backoff:  "invalid",
				},
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for invalid retries")
	}

	if !strings.Contains(err.Error(), "attempts") || !strings.Contains(err.Error(), "backoff") {
		t.Errorf("Expected retries errors, got: %v", err)
	}
}

func TestValidate_UndefinedSecret(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Env: &Environment{
			Secrets: []string{"DATABASE_URL"},
		},
		Setup: []Step{
			{
				Name:           "Use undefined secret",
				Run:            "echo hello",
				EnvFromSecrets: []string{"UNDEFINED_SECRET"},
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for undefined secret")
	}

	if !strings.Contains(err.Error(), "not declared") {
		t.Errorf("Expected secret declaration error, got: %v", err)
	}
}

func TestValidate_ValidSecrets(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Env: &Environment{
			Secrets: []string{"DATABASE_URL", "API_KEY"},
		},
		Setup: []Step{
			{
				Name:           "Use valid secrets",
				Run:            "echo hello",
				EnvFromSecrets: []string{"DATABASE_URL", "API_KEY"},
			},
		},
	}

	err := Validate(config)
	if err != nil {
		t.Errorf("Expected no validation errors, got: %v", err)
	}
}

func TestValidate_AbsolutePath(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "Absolute path",
				Run:  "echo hello",
				CWD:  "/absolute/path",
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for absolute path")
	}

	if !strings.Contains(err.Error(), "relative path") {
		t.Errorf("Expected relative path error, got: %v", err)
	}
}

func TestValidate_ParentDirectoryPath(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "Parent directory path",
				Run:  "echo hello",
				CWD:  "../parent",
			},
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for parent directory path")
	}

	if !strings.Contains(err.Error(), "must not contain") {
		t.Errorf("Expected parent directory error, got: %v", err)
	}
}

func TestValidate_ValidPaths(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Setup: []Step{
			{
				Name: "Valid relative path",
				Run:  "echo hello",
				CWD:  "subdirectory",
			},
			{
				Name: "Valid nested path",
				Run:  "echo hello",
				CWD:  "deep/nested/path",
			},
		},
	}

	err := Validate(config)
	if err != nil {
		t.Errorf("Expected no validation errors, got: %v", err)
	}
}

func TestValidate_DefaultTimeout(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Defaults: &Defaults{
			Timeout: "invalid-timeout",
		},
	}

	err := Validate(config)
	if err == nil {
		t.Fatal("Expected validation error for invalid default timeout")
	}

	if !strings.Contains(err.Error(), "defaults.timeout") {
		t.Errorf("Expected default timeout error, got: %v", err)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		{Field: "version", Message: "must be 1"},
		{Field: "name", Message: "is required"},
	}

	errStr := errors.Error()
	if !strings.Contains(errStr, "version: must be 1") {
		t.Errorf("Expected version error in output, got: %s", errStr)
	}
	if !strings.Contains(errStr, "name: is required") {
		t.Errorf("Expected name error in output, got: %s", errStr)
	}
	if !strings.Contains(errStr, ";") {
		t.Errorf("Expected semicolon separator, got: %s", errStr)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "version",
		Message: "must be 1",
	}

	expected := "version: must be 1"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}
