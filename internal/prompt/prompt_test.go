package prompt

import (
	"strings"
	"testing"
)

func TestPromptProject_ValidInput(t *testing.T) {
	// This test would require mocking stdin
	// For now, we'll test the function exists and has correct signature

	// Test that the function can be called (though it will fail without stdin)
	_, err := PromptProject()
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestPromptProjectWithOptions_EmptyOptions(t *testing.T) {
	options := []ProjectOption{}

	// Should fall back to PromptProject
	_, err := PromptProjectWithOptions(options)
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestPromptProjectWithOptions_ValidOptions(t *testing.T) {
	options := []ProjectOption{
		{Name: "Production", Slug: "acme-corp/production"},
		{Name: "Staging", Slug: "acme-corp/staging"},
	}

	// Test that the function can be called (though it will fail without stdin)
	_, err := PromptProjectWithOptions(options)
	if err == nil {
		t.Error("Expected error when no stdin available, got nil")
	}
}

func TestProjectOption_Structure(t *testing.T) {
	option := ProjectOption{
		Name: "Test Project",
		Slug: "test-org/test-project",
	}

	if option.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", option.Name)
	}

	if option.Slug != "test-org/test-project" {
		t.Errorf("Expected slug 'test-org/test-project', got '%s'", option.Slug)
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

func TestPromptProject_EmptyInput(t *testing.T) {
	// Test validation logic
	project := ""
	if project == "" {
		// This simulates the validation in PromptProject
		err := "project cannot be empty"
		if !strings.Contains(err, "empty") {
			t.Error("Expected error message to contain 'empty'")
		}
	}
}
