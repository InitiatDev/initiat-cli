package workspace

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/config"
)

func TestGetWorkspaceContext_WithValidFlags(t *testing.T) {
	workspacePath := "test-org/test-workspace"
	org := ""
	workspace := ""

	ctx, err := GetWorkspaceContext(workspacePath, org, workspace)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.OrgSlug != "test-org" {
		t.Errorf("Expected org slug 'test-org', got '%s'", ctx.OrgSlug)
	}

	if ctx.WorkspaceSlug != "test-workspace" {
		t.Errorf("Expected workspace slug 'test-workspace', got '%s'", ctx.WorkspaceSlug)
	}
}

func TestGetWorkspaceContext_WithOrgAndWorkspace(t *testing.T) {
	workspacePath := ""
	org := "test-org"
	workspace := "test-workspace"

	ctx, err := GetWorkspaceContext(workspacePath, org, workspace)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.OrgSlug != "test-org" {
		t.Errorf("Expected org slug 'test-org', got '%s'", ctx.OrgSlug)
	}

	if ctx.WorkspaceSlug != "test-workspace" {
		t.Errorf("Expected workspace slug 'test-workspace', got '%s'", ctx.WorkspaceSlug)
	}
}

func TestGetWorkspaceContext_WithWorkspaceOnly(t *testing.T) {
	workspacePath := ""
	org := ""
	workspace := "test-workspace"

	_, err := GetWorkspaceContext(workspacePath, org, workspace)
	if err == nil {
		t.Error("Expected error when workspace is specified without org, got nil")
	}
}

func TestGetWorkspaceContext_WithEmptyFlags(t *testing.T) {
	workspacePath := ""
	org := ""
	workspace := ""

	_, err := GetWorkspaceContext(workspacePath, org, workspace)
	if err == nil {
		t.Error("Expected error when no workspace context provided, got nil")
	}
}

func TestWorkspaceContextString(t *testing.T) {
	ctx := &config.WorkspaceContext{
		OrgSlug:       "test-org",
		WorkspaceSlug: "test-workspace",
	}

	expected := "test-org/test-workspace"
	if ctx.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, ctx.String())
	}
}
