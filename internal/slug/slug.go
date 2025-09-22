package slug

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	compositeSlugParts = 2 // org-slug and workspace-slug
)

// CompositeSlug represents a parsed org-slug/workspace-slug combination
type CompositeSlug struct {
	OrgSlug       string
	WorkspaceSlug string
}

// String returns the composite slug in org-slug/workspace-slug format
func (cs CompositeSlug) String() string {
	return fmt.Sprintf("%s/%s", cs.OrgSlug, cs.WorkspaceSlug)
}

// IsEmpty returns true if either org or workspace slug is empty
func (cs CompositeSlug) IsEmpty() bool {
	return cs.OrgSlug == "" || cs.WorkspaceSlug == ""
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidateSlug validates that a slug matches the required pattern
func ValidateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug '%s' must contain only lowercase letters, numbers, and hyphens", slug)
	}
	return nil
}

// ParseCompositeSlug parses a composite slug in the format "org-slug/workspace-slug"
// Returns an error if the format is invalid or slugs don't match the pattern
func ParseCompositeSlug(compositeSlug string) (CompositeSlug, error) {
	if compositeSlug == "" {
		return CompositeSlug{}, fmt.Errorf("composite slug cannot be empty")
	}

	parts := strings.Split(compositeSlug, "/")
	if len(parts) != compositeSlugParts {
		return CompositeSlug{}, fmt.Errorf(
			"composite slug must be in format 'org-slug/workspace-slug', got '%s'", compositeSlug)
	}

	orgSlug := strings.TrimSpace(parts[0])
	workspaceSlug := strings.TrimSpace(parts[1])

	if err := ValidateSlug(orgSlug); err != nil {
		return CompositeSlug{}, fmt.Errorf("invalid organization slug: %w", err)
	}

	if err := ValidateSlug(workspaceSlug); err != nil {
		return CompositeSlug{}, fmt.Errorf("invalid workspace slug: %w", err)
	}

	return CompositeSlug{
		OrgSlug:       orgSlug,
		WorkspaceSlug: workspaceSlug,
	}, nil
}

// ResolveWorkspaceSlug resolves a workspace identifier that could be:
// 1. A composite slug: "org-slug/workspace-slug"
// 2. A workspace slug only: "workspace-slug" (requires default org context)
// Returns the resolved composite slug or an error
func ResolveWorkspaceSlug(input string, defaultOrgSlug string) (CompositeSlug, error) {
	if input == "" {
		return CompositeSlug{}, fmt.Errorf("workspace identifier cannot be empty")
	}

	// Check if input contains a slash (composite slug format)
	if strings.Contains(input, "/") {
		return ParseCompositeSlug(input)
	}

	// Single slug - requires default organization context
	if defaultOrgSlug == "" {
		return CompositeSlug{}, fmt.Errorf(
			"workspace slug '%s' requires organization context. "+
				"Use 'org-slug/workspace-slug' format or set default organization with 'initiat config use org <org-slug>'",
			input)
	}

	// Validate the workspace slug
	if err := ValidateSlug(input); err != nil {
		return CompositeSlug{}, fmt.Errorf("invalid workspace slug: %w", err)
	}

	// Validate the default org slug
	if err := ValidateSlug(defaultOrgSlug); err != nil {
		return CompositeSlug{}, fmt.Errorf("invalid default organization slug: %w", err)
	}

	return CompositeSlug{
		OrgSlug:       defaultOrgSlug,
		WorkspaceSlug: input,
	}, nil
}

// BuildCompositeSlug creates a composite slug from separate org and workspace slugs
func BuildCompositeSlug(orgSlug, workspaceSlug string) (CompositeSlug, error) {
	if err := ValidateSlug(orgSlug); err != nil {
		return CompositeSlug{}, fmt.Errorf("invalid organization slug: %w", err)
	}

	if err := ValidateSlug(workspaceSlug); err != nil {
		return CompositeSlug{}, fmt.Errorf("invalid workspace slug: %w", err)
	}

	return CompositeSlug{
		OrgSlug:       orgSlug,
		WorkspaceSlug: workspaceSlug,
	}, nil
}
