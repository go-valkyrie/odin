/**
 * MIT License
 *
 * Copyright (c) 2025 Stefan Nuxoll
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

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
