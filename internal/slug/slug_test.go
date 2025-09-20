package slug

import (
	"testing"
)

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{
			name:    "valid slug with letters and numbers",
			slug:    "acme-corp-123",
			wantErr: false,
		},
		{
			name:    "valid slug with only letters",
			slug:    "production",
			wantErr: false,
		},
		{
			name:    "valid slug with only numbers",
			slug:    "123",
			wantErr: false,
		},
		{
			name:    "valid slug with hyphens",
			slug:    "my-org-name",
			wantErr: false,
		},
		{
			name:    "empty slug",
			slug:    "",
			wantErr: true,
		},
		{
			name:    "slug with uppercase letters",
			slug:    "Acme-Corp",
			wantErr: true,
		},
		{
			name:    "slug with underscores",
			slug:    "acme_corp",
			wantErr: true,
		},
		{
			name:    "slug with spaces",
			slug:    "acme corp",
			wantErr: true,
		},
		{
			name:    "slug with special characters",
			slug:    "acme@corp",
			wantErr: true,
		},
		{
			name:    "slug starting with hyphen",
			slug:    "-acme",
			wantErr: true,
		},
		{
			name:    "slug ending with hyphen",
			slug:    "acme-",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlug(tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSlug() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseCompositeSlug(t *testing.T) {
	tests := []struct {
		name          string
		compositeSlug string
		want          CompositeSlug
		wantErr       bool
	}{
		{
			name:          "valid composite slug",
			compositeSlug: "acme-corp/production",
			want: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			wantErr: false,
		},
		{
			name:          "valid composite slug with numbers",
			compositeSlug: "org123/workspace456",
			want: CompositeSlug{
				OrgSlug:       "org123",
				WorkspaceSlug: "workspace456",
			},
			wantErr: false,
		},
		{
			name:          "empty composite slug",
			compositeSlug: "",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "missing workspace slug",
			compositeSlug: "acme-corp/",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "missing org slug",
			compositeSlug: "/production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "no slash separator",
			compositeSlug: "acme-corp-production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "multiple slashes",
			compositeSlug: "acme/corp/production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "invalid org slug",
			compositeSlug: "Acme-Corp/production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "invalid workspace slug",
			compositeSlug: "acme-corp/Production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "whitespace handling",
			compositeSlug: " acme-corp / production ",
			want: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCompositeSlug(tt.compositeSlug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompositeSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCompositeSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveWorkspaceSlug(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		defaultOrgSlug string
		want           CompositeSlug
		wantErr        bool
	}{
		{
			name:           "composite slug with default org",
			input:          "acme-corp/production",
			defaultOrgSlug: "default-org",
			want: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			wantErr: false,
		},
		{
			name:           "composite slug without default org",
			input:          "acme-corp/production",
			defaultOrgSlug: "",
			want: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			wantErr: false,
		},
		{
			name:           "workspace slug only with default org",
			input:          "production",
			defaultOrgSlug: "acme-corp",
			want: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			wantErr: false,
		},
		{
			name:           "workspace slug only without default org",
			input:          "production",
			defaultOrgSlug: "",
			want:           CompositeSlug{},
			wantErr:        true,
		},
		{
			name:           "empty input",
			input:          "",
			defaultOrgSlug: "acme-corp",
			want:           CompositeSlug{},
			wantErr:        true,
		},
		{
			name:           "invalid workspace slug",
			input:          "Production",
			defaultOrgSlug: "acme-corp",
			want:           CompositeSlug{},
			wantErr:        true,
		},
		{
			name:           "invalid default org slug",
			input:          "production",
			defaultOrgSlug: "Acme-Corp",
			want:           CompositeSlug{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveWorkspaceSlug(tt.input, tt.defaultOrgSlug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveWorkspaceSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveWorkspaceSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompositeSlug_String(t *testing.T) {
	tests := []struct {
		name string
		cs   CompositeSlug
		want string
	}{
		{
			name: "normal composite slug",
			cs: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			want: "acme-corp/production",
		},
		{
			name: "empty composite slug",
			cs:   CompositeSlug{},
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.String(); got != tt.want {
				t.Errorf("CompositeSlug.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompositeSlug_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		cs   CompositeSlug
		want bool
	}{
		{
			name: "both slugs present",
			cs: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			want: false,
		},
		{
			name: "missing org slug",
			cs: CompositeSlug{
				OrgSlug:       "",
				WorkspaceSlug: "production",
			},
			want: true,
		},
		{
			name: "missing workspace slug",
			cs: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "",
			},
			want: true,
		},
		{
			name: "both slugs empty",
			cs:   CompositeSlug{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.IsEmpty(); got != tt.want {
				t.Errorf("CompositeSlug.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildCompositeSlug(t *testing.T) {
	tests := []struct {
		name          string
		orgSlug       string
		workspaceSlug string
		want          CompositeSlug
		wantErr       bool
	}{
		{
			name:          "valid slugs",
			orgSlug:       "acme-corp",
			workspaceSlug: "production",
			want: CompositeSlug{
				OrgSlug:       "acme-corp",
				WorkspaceSlug: "production",
			},
			wantErr: false,
		},
		{
			name:          "invalid org slug",
			orgSlug:       "Acme-Corp",
			workspaceSlug: "production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "invalid workspace slug",
			orgSlug:       "acme-corp",
			workspaceSlug: "Production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "empty org slug",
			orgSlug:       "",
			workspaceSlug: "production",
			want:          CompositeSlug{},
			wantErr:       true,
		},
		{
			name:          "empty workspace slug",
			orgSlug:       "acme-corp",
			workspaceSlug: "",
			want:          CompositeSlug{},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildCompositeSlug(tt.orgSlug, tt.workspaceSlug)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCompositeSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildCompositeSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}
