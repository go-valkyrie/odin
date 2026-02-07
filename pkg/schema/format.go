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
	"io"
	"strings"

	"github.com/fatih/color"
)

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

// FormatDeclarations writes declarations grouped by category to w in terminal format.
func FormatDeclarations(w io.Writer, declarations []*Declaration, indent int) {
	// Group declarations by category
	var refs, exts, others []*Declaration
	for _, d := range declarations {
		switch d.Category {
		case DeclarationRef:
			refs = append(refs, d)
		case DeclarationExt:
			exts = append(exts, d)
		case DeclarationOther:
			others = append(others, d)
		}
	}

	// Format each category group
	if len(refs) > 0 {
		header := color.New(color.Bold, color.FgCyan).SprintFunc()
		fmt.Fprintln(w)
		fmt.Fprintln(w, header("References:"))
		formatDeclarationGroup(w, refs, indent)
	}

	if len(exts) > 0 {
		header := color.New(color.Bold, color.FgCyan).SprintFunc()
		fmt.Fprintln(w)
		fmt.Fprintln(w, header("Extensions:"))
		formatDeclarationGroup(w, exts, indent)
	}

	if len(others) > 0 {
		header := color.New(color.Bold, color.FgCyan).SprintFunc()
		fmt.Fprintln(w)
		fmt.Fprintln(w, header("Declarations:"))
		formatDeclarationGroup(w, others, indent)
	}
}

func formatDeclarationGroup(w io.Writer, declarations []*Declaration, indent int) {
	for _, d := range declarations {
		prefix := strings.Repeat(" ", indent)

		// Doc comments always go above the declaration
		if d.Doc != "" {
			for _, line := range strings.Split(d.Doc, "\n") {
				fmt.Fprintf(w, "%s%s %s\n", prefix, commentMark("//"), commentText(line))
			}
		}

		if len(d.Children) > 0 {
			fmt.Fprintf(w, "%s%s\n", prefix, fieldName(d.Name))
			FormatSchema(w, d.Children, indent+2)
		} else {
			// Pad the name to at least 20 chars for alignment
			padding := 20 - len(d.Name)
			if padding < 1 {
				padding = 1
			}
			fmt.Fprintf(w, "%s%s%s%s\n", prefix, fieldName(d.Name), strings.Repeat(" ", padding), typeName(d.Type))
		}
	}
}

// FormatDeclarationsMarkdown writes declarations grouped by category to w in markdown format.
func FormatDeclarationsMarkdown(w io.Writer, declarations []*Declaration, depth int) {
	// Group declarations by category
	var refs, exts, others []*Declaration
	for _, d := range declarations {
		switch d.Category {
		case DeclarationRef:
			refs = append(refs, d)
		case DeclarationExt:
			exts = append(exts, d)
		case DeclarationOther:
			others = append(others, d)
		}
	}

	// Format each category group
	if len(refs) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## References")
		fmt.Fprintln(w)
		formatDeclarationGroupMarkdown(w, refs, depth)
	}

	if len(exts) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## Extensions")
		fmt.Fprintln(w)
		formatDeclarationGroupMarkdown(w, exts, depth)
	}

	if len(others) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## Declarations")
		fmt.Fprintln(w)
		formatDeclarationGroupMarkdown(w, others, depth)
	}
}

func formatDeclarationGroupMarkdown(w io.Writer, declarations []*Declaration, depth int) {
	for _, d := range declarations {
		// Print doc comments
		if d.Doc != "" {
			for _, line := range strings.Split(d.Doc, "\n") {
				fmt.Fprintln(w, line)
			}
			fmt.Fprintln(w)
		}

		if len(d.Children) > 0 {
			// Declaration with struct type: name followed by nested children
			fmt.Fprintf(w, "- **%s**\n", d.Name)
			FormatSchemaMarkdown(w, d.Children, depth+1)
		} else {
			// Leaf declaration: name with type
			fmt.Fprintf(w, "- **%s**: `%s`\n", d.Name, d.Type)
		}
	}
}
