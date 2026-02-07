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

package schema

import (
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

func TestWalkDeclarations(t *testing.T) {
	src := `
package test

#Component: {
	config: {
		name: string
	}

	// A reference definition
	#refs: {
		@odin(ref)
		foo: string
	}

	// An extension definition
	#exts: {
		@odin(ext)
		bar: string
	}

	// A regular declaration
	#other: {
		@odin(other)
		baz: string
	}

	// Hidden declaration (no @odin attribute)
	#hidden: {
		secret: string
	}
}
`

	ctx := cuecontext.New()
	value := ctx.CompileString(src)
	if value.Err() != nil {
		t.Fatalf("failed to compile source: %v", value.Err())
	}

	componentValue := value.LookupPath(cue.ParsePath("#Component"))
	if componentValue.Err() != nil {
		t.Fatalf("failed to lookup #Component: %v", componentValue.Err())
	}

	declarations := WalkDeclarations(componentValue)

	// Should have 3 declarations (refs, exts, other) but not hidden
	if len(declarations) != 3 {
		t.Errorf("expected 3 declarations, got %d", len(declarations))
		for _, d := range declarations {
			t.Logf("  found: %s (category: %s)", d.Name, d.Category)
		}
	}

	// Check categories
	categoryCount := map[DeclarationCategory]int{}
	for _, d := range declarations {
		categoryCount[d.Category]++
	}

	if categoryCount[DeclarationRef] != 1 {
		t.Errorf("expected 1 ref declaration, got %d", categoryCount[DeclarationRef])
	}
	if categoryCount[DeclarationExt] != 1 {
		t.Errorf("expected 1 ext declaration, got %d", categoryCount[DeclarationExt])
	}
	if categoryCount[DeclarationOther] != 1 {
		t.Errorf("expected 1 other declaration, got %d", categoryCount[DeclarationOther])
	}
}

func TestOdinHidden(t *testing.T) {
	src := `
package test

#Component: {
	config: {
		// Visible field
		name: string

		// Hidden field
		secret: string @odin(hidden)

		// Nested struct with hidden field
		database: {
			host: string
			password: string @odin(hidden)
		}
	}

	// Hidden declaration
	#internal: {
		@odin(hidden)
		internalField: string
	}

	// Visible declaration
	#refs: {
		@odin(ref)
		refField: string
	}
}
`

	ctx := cuecontext.New()
	value := ctx.CompileString(src)
	if value.Err() != nil {
		t.Fatalf("failed to compile source: %v", value.Err())
	}

	componentValue := value.LookupPath(cue.ParsePath("#Component"))
	if componentValue.Err() != nil {
		t.Fatalf("failed to lookup #Component: %v", componentValue.Err())
	}

	t.Run("WalkSchema filters hidden fields", func(t *testing.T) {
		configValue := componentValue.LookupPath(cue.ParsePath("config"))
		if configValue.Err() != nil {
			t.Fatalf("failed to lookup config: %v", configValue.Err())
		}

		fields := WalkSchema(configValue)

		// Should have 2 fields: name and database (secret should be filtered)
		if len(fields) != 2 {
			t.Errorf("expected 2 fields, got %d", len(fields))
			for _, f := range fields {
				t.Logf("  found: %s", f.Name)
			}
		}

		// Verify name field exists
		foundName := false
		for _, f := range fields {
			if f.Name == "name" {
				foundName = true
			}
			if f.Name == "secret" {
				t.Errorf("secret field should be hidden but was found")
			}
		}
		if !foundName {
			t.Errorf("name field not found")
		}

		// Check nested database struct
		var databaseField *SchemaField
		for _, f := range fields {
			if f.Name == "database" {
				databaseField = f
				break
			}
		}
		if databaseField == nil {
			t.Fatalf("database field not found")
		}

		// Database should have only host (password should be filtered)
		if len(databaseField.Children) != 1 {
			t.Errorf("expected 1 child in database, got %d", len(databaseField.Children))
			for _, c := range databaseField.Children {
				t.Logf("  found: %s", c.Name)
			}
		}

		foundHost := false
		for _, c := range databaseField.Children {
			if c.Name == "host" {
				foundHost = true
			}
			if c.Name == "password" {
				t.Errorf("password field should be hidden but was found")
			}
		}
		if !foundHost {
			t.Errorf("host field not found in database")
		}
	})

	t.Run("WalkDeclarations filters hidden declarations", func(t *testing.T) {
		declarations := WalkDeclarations(componentValue)

		// Should have only #refs (not #internal)
		if len(declarations) != 1 {
			t.Errorf("expected 1 declaration, got %d", len(declarations))
			for _, d := range declarations {
				t.Logf("  found: %s (category: %s)", d.Name, d.Category)
			}
		}

		// Verify #refs exists
		foundRefs := false
		for _, d := range declarations {
			if d.Name == "#refs" {
				foundRefs = true
			}
			if d.Name == "#internal" {
				t.Errorf("#internal declaration should be hidden but was found")
			}
		}
		if !foundRefs {
			t.Errorf("#refs declaration not found")
		}
	})
}

func TestDeclarationWithPrivateDefinitionRef(t *testing.T) {
	src := `
package test

// Private definition for references
_#SecretStoreRef: {
	name: string
	kind: string | *"SecretStore"
}

#Component: {
	// Declaration using private definition reference with unification
	#ref: _#SecretStoreRef & {
		@odin(ref)
		name: "default"
	}

	// Config field using private definition reference
	config: {
		store: _#SecretStoreRef
	}
}
`

	ctx := cuecontext.New()
	value := ctx.CompileString(src)
	if value.Err() != nil {
		t.Fatalf("failed to compile source: %v", value.Err())
	}

	componentValue := value.LookupPath(cue.ParsePath("#Component"))
	if componentValue.Err() != nil {
		t.Fatalf("failed to lookup #Component: %v", componentValue.Err())
	}

	t.Run("declaration with expand=false shows definition name", func(t *testing.T) {
		declarations := WalkDeclarations(componentValue, WithExpand(false))

		if len(declarations) != 1 {
			t.Fatalf("expected 1 declaration, got %d", len(declarations))
		}

		decl := declarations[0]
		if decl.Name != "#ref" {
			t.Errorf("expected declaration name #ref, got %s", decl.Name)
		}

		if decl.Type != "_#SecretStoreRef" {
			t.Errorf("expected type _#SecretStoreRef, got %s", decl.Type)
		}

		if len(decl.Children) > 0 {
			t.Errorf("expected no children when expand=false, got %d children", len(decl.Children))
		}
	})

	t.Run("declaration with expand=true shows expanded children", func(t *testing.T) {
		declarations := WalkDeclarations(componentValue, WithExpand(true))

		if len(declarations) != 1 {
			t.Fatalf("expected 1 declaration, got %d", len(declarations))
		}

		decl := declarations[0]
		if decl.Name != "#ref" {
			t.Errorf("expected declaration name #ref, got %s", decl.Name)
		}

		if len(decl.Children) == 0 {
			t.Errorf("expected children when expand=true, got none")
		}

		// Verify we have name and kind fields
		foundName := false
		foundKind := false
		for _, child := range decl.Children {
			if child.Name == "name" {
				foundName = true
			}
			if child.Name == "kind" {
				foundKind = true
			}
		}

		if !foundName {
			t.Errorf("expected to find 'name' field in expanded children")
		}
		if !foundKind {
			t.Errorf("expected to find 'kind' field in expanded children")
		}
	})

	t.Run("config field shows definition name when expand=false", func(t *testing.T) {
		configValue := componentValue.LookupPath(cue.ParsePath("config"))
		if configValue.Err() != nil {
			t.Fatalf("failed to lookup config: %v", configValue.Err())
		}

		fields := WalkSchema(configValue, WithExpand(false))

		if len(fields) != 1 {
			t.Fatalf("expected 1 field, got %d", len(fields))
		}

		field := fields[0]
		if field.Name != "store" {
			t.Errorf("expected field name store, got %s", field.Name)
		}

		if field.Type != "_#SecretStoreRef" {
			t.Errorf("expected type _#SecretStoreRef, got %s", field.Type)
		}

		if len(field.Children) > 0 {
			t.Errorf("expected no children when expand=false, got %d children", len(field.Children))
		}
	})
}

func TestOdinExpandAttribute(t *testing.T) {
	src := `
package test

// Private definition for references
_#SecretStoreRef: {
	name: string
	kind: string | *"SecretStore"
}

#Component: {
	config: {
		// Regular field without @odin(expand) - should not expand by default
		regularRef: _#SecretStoreRef

		// Field with @odin(expand) - should always expand
		forcedRef: _#SecretStoreRef @odin(expand)
	}

	// Declaration without @odin(expand)
	#regularDecl: _#SecretStoreRef & {
		@odin(ref)
		name: "regular"
	}

	// Declaration with @odin(expand)
	#forcedDecl: _#SecretStoreRef & {
		@odin(ref)
		@odin(expand)
		name: "forced"
	}
}
`

	ctx := cuecontext.New()
	value := ctx.CompileString(src)
	if value.Err() != nil {
		t.Fatalf("failed to compile source: %v", value.Err())
	}

	componentValue := value.LookupPath(cue.ParsePath("#Component"))
	if componentValue.Err() != nil {
		t.Fatalf("failed to lookup #Component: %v", componentValue.Err())
	}

	t.Run("config field without @odin(expand) shows type name when expand=false", func(t *testing.T) {
		configValue := componentValue.LookupPath(cue.ParsePath("config"))
		if configValue.Err() != nil {
			t.Fatalf("failed to lookup config: %v", configValue.Err())
		}

		fields := WalkSchema(configValue, WithExpand(false))

		var regularField *SchemaField
		for _, f := range fields {
			if f.Name == "regularRef" {
				regularField = f
				break
			}
		}

		if regularField == nil {
			t.Fatalf("regularRef field not found")
		}

		if regularField.Type != "_#SecretStoreRef" {
			t.Errorf("expected type _#SecretStoreRef, got %s", regularField.Type)
		}

		if len(regularField.Children) > 0 {
			t.Errorf("expected no children, got %d", len(regularField.Children))
		}
	})

	t.Run("config field with @odin(expand) expands even when expand=false", func(t *testing.T) {
		configValue := componentValue.LookupPath(cue.ParsePath("config"))
		if configValue.Err() != nil {
			t.Fatalf("failed to lookup config: %v", configValue.Err())
		}

		fields := WalkSchema(configValue, WithExpand(false))

		var forcedField *SchemaField
		for _, f := range fields {
			if f.Name == "forcedRef" {
				forcedField = f
				break
			}
		}

		if forcedField == nil {
			t.Fatalf("forcedRef field not found")
		}

		if len(forcedField.Children) == 0 {
			t.Errorf("expected children due to @odin(expand), got none")
		}

		// Verify we have name and kind fields
		foundName := false
		foundKind := false
		for _, child := range forcedField.Children {
			if child.Name == "name" {
				foundName = true
			}
			if child.Name == "kind" {
				foundKind = true
			}
		}

		if !foundName {
			t.Errorf("expected to find 'name' field in expanded children")
		}
		if !foundKind {
			t.Errorf("expected to find 'kind' field in expanded children")
		}
	})

	t.Run("declaration without @odin(expand) shows type name when expand=false", func(t *testing.T) {
		declarations := WalkDeclarations(componentValue, WithExpand(false))

		var regularDecl *Declaration
		for _, d := range declarations {
			if d.Name == "#regularDecl" {
				regularDecl = d
				break
			}
		}

		if regularDecl == nil {
			t.Fatalf("#regularDecl not found")
		}

		if regularDecl.Type != "_#SecretStoreRef" {
			t.Errorf("expected type _#SecretStoreRef, got %s", regularDecl.Type)
		}

		if len(regularDecl.Children) > 0 {
			t.Errorf("expected no children, got %d", len(regularDecl.Children))
		}
	})

	t.Run("declaration with @odin(expand) expands even when expand=false", func(t *testing.T) {
		declarations := WalkDeclarations(componentValue, WithExpand(false))

		var forcedDecl *Declaration
		for _, d := range declarations {
			if d.Name == "#forcedDecl" {
				forcedDecl = d
				break
			}
		}

		if forcedDecl == nil {
			t.Fatalf("#forcedDecl not found")
		}

		if len(forcedDecl.Children) == 0 {
			t.Errorf("expected children due to @odin(expand), got none")
		}

		// Verify we have name and kind fields
		foundName := false
		foundKind := false
		for _, child := range forcedDecl.Children {
			if child.Name == "name" {
				foundName = true
			}
			if child.Name == "kind" {
				foundKind = true
			}
		}

		if !foundName {
			t.Errorf("expected to find 'name' field in expanded children")
		}
		if !foundKind {
			t.Errorf("expected to find 'kind' field in expanded children")
		}
	})
}
