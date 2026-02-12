// SPDX-License-Identifier: MIT

package schema_test

import (
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"go-valkyrie.com/odin/pkg/schema"
)

// TestWalkSchemaBasic verifies basic schema walking functionality.
func TestWalkSchemaBasic(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString(`
		#Config: {
			name: string
			count?: int
			enabled: bool | *true
		}
	`)

	config := v.LookupPath(cue.ParsePath("#Config"))
	fields := schema.WalkSchema(config)

	if len(fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fields))
	}

	// Verify field names
	names := make(map[string]bool)
	for _, f := range fields {
		names[f.Name] = true
	}

	if !names["name"] || !names["count"] || !names["enabled"] {
		t.Errorf("missing expected fields: %v", names)
	}
}

// TestWalkSchemaWithExpand verifies the WithExpand option.
func TestWalkSchemaWithExpand(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString(`
		#Inner: {
			value: string
		}
		#Outer: {
			inner: #Inner
		}
	`)

	outer := v.LookupPath(cue.ParsePath("#Outer"))

	// Without expand - should show definition name
	fields := schema.WalkSchema(outer)
	if len(fields) != 1 || fields[0].Type != "#Inner" {
		t.Errorf("without expand: expected type #Inner, got %s", fields[0].Type)
	}

	// With expand - should inline the definition
	fieldsExpanded := schema.WalkSchema(outer, schema.WithExpand(true))
	if len(fieldsExpanded) != 1 {
		t.Fatalf("with expand: expected 1 field, got %d", len(fieldsExpanded))
	}
	if len(fieldsExpanded[0].Children) != 1 {
		t.Errorf("with expand: expected 1 child, got %d", len(fieldsExpanded[0].Children))
	}
}
