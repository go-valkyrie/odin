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
	"fmt"
	"strings"

	"cuelang.org/go/cue"
)

// SchemaField represents a single field in a CUE schema tree.
type SchemaField struct {
	Name      string
	Doc       string
	Type      string
	Optional  bool
	Required  bool
	IsPattern bool
	Default   string
	Children  []*SchemaField
}

// DeclarationCategory represents the category of a declaration based on @odin attribute.
type DeclarationCategory string

const (
	DeclarationRef   DeclarationCategory = "ref"   // @odin(ref)
	DeclarationExt   DeclarationCategory = "ext"   // @odin(ext)
	DeclarationOther DeclarationCategory = "other" // @odin or @odin(other)
)

// Declaration represents a root-level CUE definition annotated with @odin.
type Declaration struct {
	Name     string
	Doc      string
	Category DeclarationCategory
	Type     string
	Children []*SchemaField
}

// walkOptions holds options for WalkSchema.
type walkOptions struct {
	expand bool
}

// WalkOption is a functional option for WalkSchema.
type WalkOption func(*walkOptions)

// WithExpand controls whether referenced definitions are inlined (true)
// or shown by name (false).
func WithExpand(expand bool) WalkOption {
	return func(o *walkOptions) {
		o.expand = expand
	}
}

// hasOdinHidden checks if a value has @odin(hidden) attribute.
func hasOdinHidden(v cue.Value) bool {
	attrs := v.Attributes(cue.ValueAttr)
	for _, a := range attrs {
		if a.Name() == "odin" {
			if arg, err := a.String(0); err == nil && arg == "hidden" {
				return true
			}
		}
	}
	return false
}

// WalkSchema traverses a cue.Value's schema tree and returns a tree of SchemaField.
// Options can be provided to control behavior (e.g., WithExpand).
func WalkSchema(value cue.Value, opts ...WalkOption) []*SchemaField {
	o := &walkOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return walkFields(value, o.expand)
}

func walkFields(value cue.Value, expand bool) []*SchemaField {
	iter, err := value.Fields(cue.Optional(true))
	if err != nil {
		return nil
	}

	var fields []*SchemaField
	for iter.Next() {
		// Skip fields with @odin(hidden) attribute
		if hasOdinHidden(iter.Value()) {
			continue
		}
		f := fieldFromIter(iter, expand)
		fields = append(fields, f)
	}

	// Also walk pattern constraints
	iter, err = value.Fields(cue.Patterns(true))
	if err == nil {
		for iter.Next() {
			sel := iter.Selector()
			if sel.ConstraintType() == cue.PatternConstraint {
				// Skip pattern constraints with @odin(hidden) attribute
				if hasOdinHidden(iter.Value()) {
					continue
				}
				f := &SchemaField{
					Name:      sel.String(),
					IsPattern: true,
				}
				populateFieldValue(f, iter.Value(), expand)
				fields = append(fields, f)
			}
		}
	}

	return fields
}

func fieldFromIter(iter *cue.Iterator, expand bool) *SchemaField {
	sel := iter.Selector()
	name := sel.String()
	// Selector.String() includes optionality markers (? and !), strip them
	// since we track optionality separately
	name = strings.TrimRight(name, "?!")
	f := &SchemaField{
		Name:     name,
		Optional: iter.IsOptional(),
		Required: sel.ConstraintType() == cue.RequiredConstraint,
	}

	// Extract doc comments
	docs := iter.Value().Doc()
	var docParts []string
	for _, cg := range docs {
		docParts = append(docParts, cg.Text())
	}
	if len(docParts) > 0 {
		f.Doc = strings.TrimSpace(strings.Join(docParts, "\n"))
	}

	populateFieldValue(f, iter.Value(), expand)
	return f
}

func populateFieldValue(f *SchemaField, v cue.Value, expand bool) {
	// Check for default value
	defVal, hasDefault := v.Default()
	if hasDefault {
		f.Default = formatValue(defVal)
	}

	kind := v.IncompleteKind()

	// Check if this is a disjunction
	op, args := v.Expr()
	if op == cue.OrOp && len(args) > 0 {
		f.Type = formatDisjunction(args)
		return
	}

	// Check if this is a definition reference (unexpanded)
	if !expand && kind == cue.StructKind {
		if defName, ok := definitionRefName(v); ok {
			f.Type = defName
			return
		}
	}

	if kind == cue.StructKind {
		children := walkFields(v, expand)
		if len(children) > 0 {
			f.Children = children
			return
		}
		f.Type = "{...}"
		return
	}

	if kind == cue.ListKind {
		f.Type = formatListType(v)
		return
	}

	f.Type = formatKind(kind)
}

func formatDisjunction(args []cue.Value) string {
	var parts []string
	for _, a := range args {
		parts = append(parts, formatValue(a))
	}
	return strings.Join(parts, " | ")
}

func formatValue(v cue.Value) string {
	switch v.IncompleteKind() {
	case cue.StringKind:
		if s, err := v.String(); err == nil {
			return fmt.Sprintf("%q", s)
		}
	case cue.BoolKind:
		if b, err := v.Bool(); err == nil {
			return fmt.Sprintf("%v", b)
		}
	case cue.IntKind:
		if i, err := v.Int64(); err == nil {
			return fmt.Sprintf("%d", i)
		}
	case cue.FloatKind:
		if f, err := v.Float64(); err == nil {
			return fmt.Sprintf("%g", f)
		}
	}
	return fmt.Sprint(v)
}

func formatListType(v cue.Value) string {
	return "[...]"
}

func formatKind(k cue.Kind) string {
	switch k {
	case cue.StringKind:
		return "string"
	case cue.BoolKind:
		return "bool"
	case cue.IntKind:
		return "int"
	case cue.FloatKind:
		return "float"
	case cue.NumberKind:
		return "number"
	case cue.BytesKind:
		return "bytes"
	case cue.NullKind:
		return "null"
	default:
		return k.String()
	}
}

// definitionRefName extracts a definition reference name from a value.
// It handles both pure references (_#Foo) and unifications (_#Foo & {...}).
// Returns the definition name and true if found, empty string and false otherwise.
func definitionRefName(v cue.Value) (string, bool) {
	// Check for pure reference
	_, path := v.ReferencePath()
	if path.String() != "" {
		sel := path.Selectors()
		if len(sel) > 0 && sel[len(sel)-1].IsDefinition() {
			return sel[len(sel)-1].String(), true
		}
	}

	// Check for unification with definition reference (e.g., _#Foo & {...})
	op, args := v.Expr()
	if op == cue.AndOp {
		for _, arg := range args {
			_, argPath := arg.ReferencePath()
			if argPath.String() != "" {
				argSel := argPath.Selectors()
				if len(argSel) > 0 && argSel[len(argSel)-1].IsDefinition() {
					return argSel[len(argSel)-1].String(), true
				}
			}
		}
	}

	return "", false
}

// WalkDeclarations traverses root-level definitions annotated with @odin attribute.
// Returns declarations grouped by category. Only definitions with @odin attribute are included.
// Private definitions (prefixed with _#) are skipped.
func WalkDeclarations(value cue.Value, opts ...WalkOption) []*Declaration {
	o := &walkOptions{}
	for _, opt := range opts {
		opt(o)
	}

	iter, err := value.Fields(cue.Definitions(true))
	if err != nil {
		return nil
	}

	var declarations []*Declaration
	for iter.Next() {
		sel := iter.Selector()
		name := sel.String()

		// Skip private definitions
		if strings.HasPrefix(name, "_#") {
			continue
		}

		// Check for @odin attribute in ValueAttrs
		attrs := iter.Value().Attributes(cue.ValueAttr)
		var odinAttr cue.Attribute
		found := false
		for _, a := range attrs {
			if a.Name() == "odin" {
				odinAttr = a
				found = true
				break
			}
		}

		if !found {
			// No @odin attribute, skip this definition
			continue
		}

		// Determine category from attribute argument
		category := DeclarationOther
		if categoryStr, err := odinAttr.String(0); err == nil {
			switch categoryStr {
			case "ref":
				category = DeclarationRef
			case "ext":
				category = DeclarationExt
			case "hidden":
				// Skip hidden declarations
				continue
			default:
				category = DeclarationOther
			}
		}

		// Extract doc comments
		docs := iter.Value().Doc()
		var docParts []string
		for _, cg := range docs {
			docParts = append(docParts, cg.Text())
		}
		doc := ""
		if len(docParts) > 0 {
			doc = strings.TrimSpace(strings.Join(docParts, "\n"))
		}

		// Create declaration
		decl := &Declaration{
			Name:     name,
			Doc:      doc,
			Category: category,
		}

		// Populate type and children using same logic as populateFieldValue
		v := iter.Value()
		kind := v.IncompleteKind()

		if kind == cue.StructKind {
			// Check if this is a definition reference when not expanding
			if !o.expand {
				if defName, ok := definitionRefName(v); ok {
					decl.Type = defName
					declarations = append(declarations, decl)
					continue
				}
			}

			children := walkFields(v, o.expand)
			if len(children) > 0 {
				decl.Children = children
				decl.Type = "{...}"
			} else {
				decl.Type = "{...}"
			}
		} else if kind == cue.ListKind {
			decl.Type = "[...]"
		} else {
			decl.Type = formatKind(kind)
		}

		declarations = append(declarations, decl)
	}

	return declarations
}
