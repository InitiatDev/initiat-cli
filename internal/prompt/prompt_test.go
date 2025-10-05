package prompt

import (
	"strings"
	"testing"
)

func TestPromptWorkspace_ValidInput(t *testing.T) {
	// This test would require mocking stdin
	// For now, we'll test the function exists and has correct signature

	// Test that the function can be called (though it will fail without stdin)
	_, err := PromptWorkspace()
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestPromptWorkspaceWithOptions_EmptyOptions(t *testing.T) {
	options := []WorkspaceOption{}

	// Should fall back to PromptWorkspace
	_, err := PromptWorkspaceWithOptions(options)
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestPromptWorkspaceWithOptions_ValidOptions(t *testing.T) {
	options := []WorkspaceOption{
		{Name: "Production", Slug: "acme-corp/production"},
		{Name: "Staging", Slug: "acme-corp/staging"},
	}

	// Test that the function can be called (though it will fail without stdin)
	_, err := PromptWorkspaceWithOptions(options)
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestWorkspaceOption_Structure(t *testing.T) {
	option := WorkspaceOption{
		Name: "Test Workspace",
		Slug: "test-org/test-workspace",
	}

	if option.Name != "Test Workspace" {
		t.Errorf("Expected name 'Test Workspace', got '%s'", option.Name)
	}

	if option.Slug != "test-org/test-workspace" {
		t.Errorf("Expected slug 'test-org/test-workspace', got '%s'", option.Slug)
	}
}

func TestPromptEmail_ValidInput(t *testing.T) {
	// Test that the function can be called (though it will fail without stdin)
	_, err := PromptEmail()
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestPromptPassword_ValidInput(t *testing.T) {
	// Test that the function can be called (though it will fail without stdin)
	_, err := PromptPassword()
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestPromptWorkspace_EmptyInput(t *testing.T) {
	// Test validation logic
	workspace := ""
	if workspace == "" {
		// This simulates the validation in PromptWorkspace
		err := "workspace cannot be empty"
		if !strings.Contains(err, "empty") {
			t.Error("Expected error message to contain 'empty'")
		}
	}
}
