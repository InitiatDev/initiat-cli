package project

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
)

func TestGetProjectContext_WithValidFlags(t *testing.T) {
	projectPath := "test-org/test-project"
	org := ""
	project := ""

	ctx, err := GetProjectContext(projectPath, org, project)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.OrgSlug != "test-org" {
		t.Errorf("Expected org slug 'test-org', got '%s'", ctx.OrgSlug)
	}

	if ctx.ProjectSlug != "test-project" {
		t.Errorf("Expected project slug 'test-project', got '%s'", ctx.ProjectSlug)
	}
}

func TestGetProjectContext_WithOrgAndProject(t *testing.T) {
	projectPath := ""
	org := "test-org"
	project := "test-project"

	ctx, err := GetProjectContext(projectPath, org, project)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.OrgSlug != "test-org" {
		t.Errorf("Expected org slug 'test-org', got '%s'", ctx.OrgSlug)
	}

	if ctx.ProjectSlug != "test-project" {
		t.Errorf("Expected project slug 'test-project', got '%s'", ctx.ProjectSlug)
	}
}

func TestGetProjectContext_WithProjectOnly(t *testing.T) {
	projectPath := ""
	org := ""
	project := "test-project"

	_, err := GetProjectContext(projectPath, org, project)
	if err == nil {
		t.Error("Expected error when project is specified without org, got nil")
	}
}

func TestGetProjectContext_WithEmptyFlags(t *testing.T) {
	projectPath := ""
	org := ""
	project := ""

	_, err := GetProjectContext(projectPath, org, project)
	if err == nil {
		t.Error("Expected error when no project context provided, got nil")
	}
}

func TestProjectContextString(t *testing.T) {
	ctx := &config.ProjectContext{
		OrgSlug:     "test-org",
		ProjectSlug: "test-project",
	}

	expected := "test-org/test-project"
	if ctx.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, ctx.String())
	}
}
