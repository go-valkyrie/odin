// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatSchemaMarkdown(t *testing.T) {
	tests := []struct {
		name         string
		fields       []*SchemaField
		depth        int
		wantContains []string
	}{
		{
			name: "simple field with type",
			fields: []*SchemaField{
				{Name: "name", Type: "string", Required: true},
			},
			depth: 0,
			wantContains: []string{
				"- **name** (required): `string`",
			},
		},
		{
			name: "optional field",
			fields: []*SchemaField{
				{Name: "count", Type: "int", Optional: true},
			},
			depth: 0,
			wantContains: []string{
				"- **count** (optional): `int`",
			},
		},
		{
			name: "field with default",
			fields: []*SchemaField{
				{Name: "enabled", Type: "bool", Default: "true"},
			},
			depth: 0,
			wantContains: []string{
				"- **enabled**: `bool` (default: true)",
			},
		},
		{
			name: "field with doc comment",
			fields: []*SchemaField{
				{Name: "name", Type: "string", Doc: "The user's name"},
			},
			depth: 0,
			wantContains: []string{
				"The user's name",
				"- **name**: `string`",
			},
		},
		{
			name: "nested struct",
			fields: []*SchemaField{
				{
					Name: "config",
					Children: []*SchemaField{
						{Name: "host", Type: "string"},
						{Name: "port", Type: "int"},
					},
				},
			},
			depth: 0,
			wantContains: []string{
				"- **config**",
				"  - **host**: `string`",
				"  - **port**: `int`",
			},
		},
		{
			name: "pattern constraint",
			fields: []*SchemaField{
				{Name: "[string]", Type: "any", IsPattern: true},
			},
			depth: 0,
			wantContains: []string{
				"- **[string]**: `any`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			FormatSchemaMarkdown(&buf, tt.fields, tt.depth)
			got := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("FormatSchemaMarkdown() output missing %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestFormatDeclarationsMarkdown(t *testing.T) {
	tests := []struct {
		name         string
		declarations []*Declaration
		depth        int
		wantContains []string
	}{
		{
			name: "reference declarations",
			declarations: []*Declaration{
				{
					Name:     "serviceAccount",
					Type:     "#ServiceAccountRef",
					Category: DeclarationRef,
					Doc:      "Reference to a service account",
				},
			},
			depth: 0,
			wantContains: []string{
				"## References",
				"Reference to a service account",
				"- **serviceAccount**: `#ServiceAccountRef`",
			},
		},
		{
			name: "extension declarations",
			declarations: []*Declaration{
				{
					Name:     "volumeMounts",
					Type:     "[...#VolumeMount]",
					Category: DeclarationExt,
					Doc:      "Additional volume mounts",
				},
			},
			depth: 0,
			wantContains: []string{
				"## Extensions",
				"Additional volume mounts",
				"- **volumeMounts**: `[...#VolumeMount]`",
			},
		},
		{
			name: "other declarations",
			declarations: []*Declaration{
				{
					Name:     "customField",
					Type:     "string",
					Category: DeclarationOther,
				},
			},
			depth: 0,
			wantContains: []string{
				"## Declarations",
				"- **customField**: `string`",
			},
		},
		{
			name: "struct declaration with children",
			declarations: []*Declaration{
				{
					Name:     "metadata",
					Category: DeclarationOther,
					Children: []*SchemaField{
						{Name: "name", Type: "string", Required: true},
						{Name: "namespace", Type: "string", Optional: true},
					},
				},
			},
			depth: 0,
			wantContains: []string{
				"## Declarations",
				"- **metadata**",
				"  - **name** (required): `string`",
				"  - **namespace** (optional): `string`",
			},
		},
		{
			name: "mixed categories",
			declarations: []*Declaration{
				{Name: "ref1", Type: "string", Category: DeclarationRef},
				{Name: "ext1", Type: "int", Category: DeclarationExt},
				{Name: "other1", Type: "bool", Category: DeclarationOther},
			},
			depth: 0,
			wantContains: []string{
				"## References",
				"- **ref1**: `string`",
				"## Extensions",
				"- **ext1**: `int`",
				"## Declarations",
				"- **other1**: `bool`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			FormatDeclarationsMarkdown(&buf, tt.declarations, tt.depth)
			got := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("FormatDeclarationsMarkdown() output missing %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}
