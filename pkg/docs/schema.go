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

package docs

import (
	"fmt"
	"io"
	"strings"

	"cuelang.org/go/cue"
	"github.com/fatih/color"
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

// WalkSchema traverses a cue.Value's schema tree and returns a tree of SchemaField.
// If expand is true, referenced definitions are inlined; otherwise their name is shown.
func WalkSchema(value cue.Value, expand bool) []*SchemaField {
	return walkFields(value, expand)
}

func walkFields(value cue.Value, expand bool) []*SchemaField {
	iter, err := value.Fields(cue.Optional(true))
	if err != nil {
		return nil
	}

	var fields []*SchemaField
	for iter.Next() {
		f := fieldFromIter(iter, expand)
		fields = append(fields, f)
	}

	// Also walk pattern constraints
	iter, err = value.Fields(cue.Patterns(true))
	if err == nil {
		for iter.Next() {
			sel := iter.Selector()
			if sel.ConstraintType() == cue.PatternConstraint {
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
		_, path := v.ReferencePath()
		if path.String() != "" {
			sel := path.Selectors()
			if len(sel) > 0 && sel[len(sel)-1].IsDefinition() {
				f.Type = sel[len(sel)-1].String()
				return
			}
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

var (
	commentMark  = color.New(color.FgHiBlack).SprintFunc()
	commentText  = color.New(color.Italic).SprintFunc()
	fieldName    = color.New(color.Bold).SprintFunc()
	typeName     = color.New(color.FgGreen).SprintFunc()
	defaultValue = color.New(color.FgYellow).SprintFunc()
)

// FormatSchema writes a human-readable schema tree to w.
func FormatSchema(w io.Writer, fields []*SchemaField, indent int) {
	for _, f := range fields {
		prefix := strings.Repeat(" ", indent)

		// Build the name with optionality markers
		name := f.Name
		if f.IsPattern {
			// Pattern constraints already have brackets
		} else if f.Required {
			name += "!"
		} else if f.Optional {
			name += "?"
		}

		// Doc comments always go above the field
		if f.Doc != "" {
			for _, line := range strings.Split(f.Doc, "\n") {
				fmt.Fprintf(w, "%s%s %s\n", prefix, commentMark("//"), commentText(line))
			}
		}

		if len(f.Children) > 0 {
			fmt.Fprintf(w, "%s%s\n", prefix, fieldName(name))
			FormatSchema(w, f.Children, indent+2)
		} else {
			typeStr := f.Type
			if f.Default != "" {
				typeStr = typeName(f.Type) + defaultValue(fmt.Sprintf(" (default: %s)", f.Default))
			} else {
				typeStr = typeName(typeStr)
			}

			// Pad the name to at least 20 chars for alignment
			padding := 20 - len(name)
			if padding < 1 {
				padding = 1
			}
			fmt.Fprintf(w, "%s%s%s%s\n", prefix, fieldName(name), strings.Repeat(" ", padding), typeStr)
		}
	}
}

// FormatSchemaMarkdown writes a schema tree to w in markdown format.
// Fields are rendered as nested lists with doc comments, types, and defaults.
func FormatSchemaMarkdown(w io.Writer, fields []*SchemaField, depth int) {
	for _, f := range fields {
		indent := strings.Repeat("  ", depth)

		// Build the name with optionality markers
		name := f.Name
		optMarker := ""
		if f.IsPattern {
			// Pattern constraints already have brackets
		} else if f.Required {
			optMarker = " (required)"
		} else if f.Optional {
			optMarker = " (optional)"
		}

		// Print doc comments before the field
		if f.Doc != "" {
			for _, line := range strings.Split(f.Doc, "\n") {
				fmt.Fprintf(w, "%s%s\n", indent, line)
			}
			fmt.Fprintln(w)
		}

		if len(f.Children) > 0 {
			// Struct field: bold name followed by nested children
			fmt.Fprintf(w, "%s- **%s**%s\n", indent, name, optMarker)
			FormatSchemaMarkdown(w, f.Children, depth+1)
		} else {
			// Leaf field: name with type and optional default
			typeInfo := fmt.Sprintf("`%s`", f.Type)
			if f.Default != "" {
				typeInfo = fmt.Sprintf("`%s` (default: %s)", f.Type, f.Default)
			}
			fmt.Fprintf(w, "%s- **%s**%s: %s\n", indent, name, optMarker, typeInfo)
		}
	}
}
