// SPDX-License-Identifier: MIT

package docs

import (
	"strings"
	"testing"

	"go-valkyrie.com/odin/pkg/model"
)

func TestResolveReference(t *testing.T) {
	templates := []*model.ComponentTemplate{
		{Package: "platform.example.com/workload", Name: "#WebApp", Module: "platform.example.com/workload", Version: "v0.1.0"},
		{Package: "platform.example.com/workload", Name: "#Deployment", Module: "platform.example.com/workload", Version: "v0.1.0"},
		{Package: "platform.example.com/security", Name: "#ServiceAccount", Module: "platform.example.com/security", Version: "v0.2.0"},
		{Package: "example.com/other", Name: "#WebApp", Module: "example.com/other", Version: "v1.0.0"}, // Duplicate definition name
		{Package: "example.com/single", Name: "#Unique", Module: "example.com/single", Version: "v1.0.0"},
	}

	tests := []struct {
		name          string
		reference     string
		templates     []*model.ComponentTemplate
		wantPackage   string
		wantName      string
		wantErrSubstr string
	}{
		// 1. Fully qualified matches
		{
			name:        "fully qualified without version",
			reference:   "platform.example.com/workload:#WebApp",
			templates:   templates,
			wantPackage: "platform.example.com/workload",
			wantName:    "#WebApp",
		},
		{
			name:        "fully qualified with version",
			reference:   "platform.example.com/workload@v0.1.0:#WebApp",
			templates:   templates,
			wantPackage: "platform.example.com/workload",
			wantName:    "#WebApp",
		},
		{
			name:          "fully qualified no match",
			reference:     "platform.example.com/workload:#Missing",
			templates:     templates,
			wantErrSubstr: "no component template found matching",
		},

		// 2. Definition name matches
		{
			name:        "definition name unique match",
			reference:   "Unique",
			templates:   templates,
			wantPackage: "example.com/single",
			wantName:    "#Unique",
		},
		{
			name:        "definition name unique match case insensitive",
			reference:   "unique",
			templates:   templates,
			wantPackage: "example.com/single",
			wantName:    "#Unique",
		},
		{
			name:          "definition name ambiguous",
			reference:     "WebApp",
			templates:     templates,
			wantErrSubstr: "ambiguous reference \"WebApp\" matches multiple templates",
		},

		// 3. Package name matches
		{
			name:        "package name unique match",
			reference:   "single",
			templates:   templates,
			wantPackage: "example.com/single",
			wantName:    "#Unique",
		},
		{
			name:        "package name unique match case insensitive",
			reference:   "SINGLE",
			templates:   templates,
			wantPackage: "example.com/single",
			wantName:    "#Unique",
		},
		{
			name:          "package name ambiguous - multiple templates in same package",
			reference:     "workload",
			templates:     templates,
			wantErrSubstr: "ambiguous reference \"workload\" matches multiple templates",
		},

		// 4. Dot syntax matches
		{
			name:        "dot syntax exact match",
			reference:   "workload.WebApp",
			templates:   templates,
			wantPackage: "platform.example.com/workload",
			wantName:    "#WebApp",
		},
		{
			name:        "dot syntax case insensitive",
			reference:   "WORKLOAD.webapp",
			templates:   templates,
			wantPackage: "platform.example.com/workload",
			wantName:    "#WebApp",
		},
		{
			name:        "dot syntax disambiguates duplicate definition names",
			reference:   "other.WebApp",
			templates:   templates,
			wantPackage: "example.com/other",
			wantName:    "#WebApp",
		},

		// 5. No match
		{
			name:          "no match",
			reference:     "NonExistent",
			templates:     templates,
			wantErrSubstr: "no component template matching \"NonExistent\"",
		},
		{
			name:          "empty templates list",
			reference:     "WebApp",
			templates:     []*model.ComponentTemplate{},
			wantErrSubstr: "no component templates available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveReference(tt.reference, tt.templates)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Package != tt.wantPackage {
				t.Errorf("Package = %q, want %q", got.Package, tt.wantPackage)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

func TestResolvePackagePath(t *testing.T) {
	templates := []*model.ComponentTemplate{
		{Package: "platform.example.com/workload", Name: "#WebApp", Module: "platform.example.com/workload", Version: "v0.1.0"},
		{Package: "platform.example.com/workload", Name: "#Deployment", Module: "platform.example.com/workload", Version: "v0.1.0"},
		{Package: "platform.example.com/security", Name: "#ServiceAccount", Module: "platform.example.com/security", Version: "v0.2.0"},
		{Package: "example.com/other", Name: "#WebApp", Module: "example.com/other", Version: "v1.0.0"},
		{Package: "platform.example.com/workload@v0.2.0", Name: "#Job", Module: "platform.example.com/workload", Version: "v0.2.0"}, // With version in Package field
	}

	tests := []struct {
		name      string
		reference string
		templates []*model.ComponentTemplate
		wantCount int
		wantFirst string // Package of first result (if any)
	}{
		{
			name:      "exact match",
			reference: "platform.example.com/workload",
			templates: templates,
			wantCount: 3, // WebApp, Deployment, Job (version stripped from both)
			wantFirst: "platform.example.com/workload",
		},
		{
			name:      "exact match with version - version stripped",
			reference: "platform.example.com/workload@v0.1.0",
			templates: templates,
			wantCount: 3, // Same as above - versions are stripped
			wantFirst: "platform.example.com/workload",
		},
		{
			name:      "prefix match",
			reference: "platform.example.com",
			templates: templates,
			wantCount: 4, // workload (3) + security (1)
			wantFirst: "platform.example.com/security",
		},
		{
			name:      "case insensitive",
			reference: "PLATFORM.EXAMPLE.COM/workload",
			templates: templates,
			wantCount: 3,
			wantFirst: "platform.example.com/workload",
		},
		{
			name:      "no match",
			reference: "nonexistent.com",
			templates: templates,
			wantCount: 0,
		},
		{
			name:      "empty templates",
			reference: "platform.example.com",
			templates: []*model.ComponentTemplate{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePackagePath(tt.reference, tt.templates)
			if len(got) != tt.wantCount {
				t.Errorf("got %d matches, want %d", len(got), tt.wantCount)
			}
			if tt.wantCount > 0 && tt.wantFirst != "" {
				if got[0].Package != tt.wantFirst {
					t.Errorf("first match Package = %q, want %q", got[0].Package, tt.wantFirst)
				}
			}
			// Verify sorting
			for i := 1; i < len(got); i++ {
				prev, curr := got[i-1], got[i]
				if prev.Package > curr.Package || (prev.Package == curr.Package && prev.Name > curr.Name) {
					t.Errorf("results not sorted: [%d]=%s.%s > [%d]=%s.%s",
						i-1, prev.Package, prev.Name, i, curr.Package, curr.Name)
				}
			}
		})
	}
}

func TestShorthandName(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{"platform.example.com/workload", "workload"},
		{"platform.example.com/workload@v0", "workload"},
		{"platform.example.com/workload@v0.1.0", "workload"},
		{"workload", "workload"},
		{"workload@v1", "workload"},
		{"example.com/foo/bar", "bar"},
		{"example.com/foo/bar@v2", "bar"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			got := shorthandName(tt.pkg)
			if got != tt.want {
				t.Errorf("shorthandName(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}
