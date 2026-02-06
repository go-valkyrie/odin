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
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"github.com/fatih/color"
	"go-valkyrie.com/odin/pkg/docs"
	"go-valkyrie.com/odin/pkg/model"
	"go-valkyrie.com/odin/pkg/schema"
)

func (o *Options) Run(ctx context.Context) error {
	return run(ctx, *o)
}

func run(ctx context.Context, opts Options) error {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	modelOpts := []model.Option{
		model.WithLogger(logger),
		model.WithRegistries(opts.Registries),
		model.WithCacheDir(opts.CacheDir),
	}

	b, err := model.LoadBundle(opts.BundlePath, modelOpts...)
	if err != nil {
		return err
	}

	var templates []*model.ComponentTemplate
	for tmpl, err := range b.ComponentTemplates(ctx) {
		if err != nil {
			return err
		}
		templates = append(templates, tmpl)
	}

	// Resolve reference to one or more templates
	var resolvedTemplates []*model.ComponentTemplate
	if strings.Contains(opts.Reference, "/") && !strings.Contains(opts.Reference, ":#") {
		// Package path reference
		resolvedTemplates = docs.ResolvePackagePath(opts.Reference, templates)
		if len(resolvedTemplates) == 0 {
			// Fall back to ResolveReference for helpful error message
			_, err := docs.ResolveReference(opts.Reference, templates)
			return err
		}
	} else {
		// Single template reference
		tmpl, err := docs.ResolveReference(opts.Reference, templates)
		if err != nil {
			return err
		}
		resolvedTemplates = []*model.ComponentTemplate{tmpl}
	}

	// Normalize format aliases
	format := opts.Format
	switch format {
	case "md":
		format = "markdown"
	case "mdm":
		format = "markdown-multi"
	case "mdb":
		format = "mdbook"
	}

	// Route to appropriate output handler
	switch format {
	case "text":
		return runTextMulti(resolvedTemplates, opts)
	case "markdown":
		return runMarkdownMulti(resolvedTemplates, opts)
	case "markdown-multi":
		return runMarkdownDirectory(resolvedTemplates, opts, false)
	case "mdbook":
		return runMarkdownDirectory(resolvedTemplates, opts, true)
	default:
		return fmt.Errorf("unsupported output format: %q (supported: text, markdown, markdown-multi, mdbook)", opts.Format)
	}
}

func runTextMulti(templates []*model.ComponentTemplate, opts Options) error {
	var w io.Writer = os.Stdout
	if opts.OutputPath != "" {
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	for i, tmpl := range templates {
		if i > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "===================================")
			fmt.Fprintln(w)
		}
		if err := runText(tmpl, opts, w); err != nil {
			return err
		}
	}
	return nil
}

func runText(tmpl *model.ComponentTemplate, opts Options, w io.Writer) error {
	header := color.New(color.Bold, color.FgCyan).SprintFunc()
	italic := color.New(color.Italic).SprintFunc()
	label := color.New(color.Bold).SprintFunc()
	value := color.New(color.FgGreen).SprintFunc()

	// Print header
	fmt.Fprintf(w, "%s %s\n", header(tmpl.Package), header(tmpl.Name))
	fmt.Fprintln(w)

	// Print doc comments
	docComments := tmpl.Value.Doc()
	for _, cg := range docComments {
		text := strings.TrimSpace(cg.Text())
		if text != "" {
			fmt.Fprintln(w, italic(text))
			fmt.Fprintln(w)
		}
	}

	// Print apiVersion and kind if available
	printConcreteField(w, tmpl.Value, "apiVersion", label, value)
	printConcreteField(w, tmpl.Value, "kind", label, value)

	// Print config schema
	fields := tmpl.ConfigSchema(schema.WithExpand(opts.Expand))
	if len(fields) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, header("Config:"))
		schema.FormatSchema(w, fields, 2)
	}

	return nil
}

func runMarkdownMulti(templates []*model.ComponentTemplate, opts Options) error {
	var w io.Writer = os.Stdout
	if opts.OutputPath != "" {
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	for i, tmpl := range templates {
		if i > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "---")
			fmt.Fprintln(w)
		}
		if err := runMarkdown(tmpl, opts, w); err != nil {
			return err
		}
	}
	return nil
}

func runMarkdown(tmpl *model.ComponentTemplate, opts Options, w io.Writer) error {
	// Print header
	fmt.Fprintf(w, "# %s %s\n\n", tmpl.Package, tmpl.Name)

	// Print doc comments as blockquote
	docComments := tmpl.Value.Doc()
	for _, cg := range docComments {
		text := strings.TrimSpace(cg.Text())
		if text != "" {
			for _, line := range strings.Split(text, "\n") {
				fmt.Fprintf(w, "> %s\n", line)
			}
			fmt.Fprintln(w)
		}
	}

	// Print apiVersion and kind in table
	apiVersion := tmpl.Value.LookupPath(cue.ParsePath("apiVersion"))
	kind := tmpl.Value.LookupPath(cue.ParsePath("kind"))
	hasApiVersion := apiVersion.Err() == nil
	hasKind := kind.Err() == nil

	if hasApiVersion || hasKind {
		fmt.Fprintln(w, "| Field | Value |")
		fmt.Fprintln(w, "|-------|-------|")
		if hasApiVersion {
			if s, err := apiVersion.String(); err == nil {
				fmt.Fprintf(w, "| apiVersion | `%s` |\n", s)
			}
		}
		if hasKind {
			if s, err := kind.String(); err == nil {
				fmt.Fprintf(w, "| kind | `%s` |\n", s)
			}
		}
		fmt.Fprintln(w)
	}

	// Print config schema
	fields := tmpl.ConfigSchema(schema.WithExpand(opts.Expand))
	if len(fields) > 0 {
		fmt.Fprintln(w, "## Config")
		fmt.Fprintln(w)
		schema.FormatSchemaMarkdown(w, fields, 0)
	}

	return nil
}

func runMarkdownDirectory(templates []*model.ComponentTemplate, opts Options, generateSummary bool) error {
	// Create output directory
	if err := os.MkdirAll(opts.OutputPath, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Group templates by package shorthand
	type pkgGroup struct {
		shorthand string
		templates []*model.ComponentTemplate
	}
	groups := make(map[string]*pkgGroup)
	var groupOrder []string

	for _, tmpl := range templates {
		shorthand := shorthandName(tmpl.Package)
		if groups[shorthand] == nil {
			groups[shorthand] = &pkgGroup{
				shorthand: shorthand,
				templates: []*model.ComponentTemplate{},
			}
			groupOrder = append(groupOrder, shorthand)
		}
		groups[shorthand].templates = append(groups[shorthand].templates, tmpl)
	}

	// Write each template to its own file
	for _, shorthand := range groupOrder {
		group := groups[shorthand]
		pkgDir := filepath.Join(opts.OutputPath, shorthand)
		if err := os.MkdirAll(pkgDir, 0755); err != nil {
			return fmt.Errorf("creating package directory: %w", err)
		}

		for _, tmpl := range group.templates {
			defName := strings.TrimPrefix(tmpl.Name, "#")
			filename := filepath.Join(pkgDir, defName+".md")
			f, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("creating file %s: %w", filename, err)
			}
			if err := runMarkdown(tmpl, opts, f); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	// Generate SUMMARY.md for mdbook format
	if generateSummary && !opts.NoSummary {
		summaryPath := filepath.Join(opts.OutputPath, "SUMMARY.md")
		f, err := os.Create(summaryPath)
		if err != nil {
			return fmt.Errorf("creating SUMMARY.md: %w", err)
		}
		defer f.Close()

		fmt.Fprintln(f, "# Summary")
		fmt.Fprintln(f)

		for _, shorthand := range groupOrder {
			group := groups[shorthand]
			if len(group.templates) == 0 {
				continue
			}

			// Chapter header (linked to first template)
			firstDefName := strings.TrimPrefix(group.templates[0].Name, "#")
			firstPath := filepath.Join(shorthand, firstDefName+".md")
			fmt.Fprintf(f, "- [%s](%s)\n", shorthand, firstPath)

			// Sub-pages for all templates
			for _, tmpl := range group.templates {
				defName := strings.TrimPrefix(tmpl.Name, "#")
				relPath := filepath.Join(shorthand, defName+".md")
				fmt.Fprintf(f, "  - [%s](%s)\n", defName, relPath)
			}
		}
	}

	return nil
}

func shorthandName(pkg string) string {
	// Strip @vN suffix if present
	if idx := strings.LastIndex(pkg, "@"); idx != -1 {
		pkg = pkg[:idx]
	}
	// Get last path segment
	if idx := strings.LastIndex(pkg, "/"); idx != -1 {
		return pkg[idx+1:]
	}
	return pkg
}

func printConcreteField(w io.Writer, v cue.Value, path string, labelFn, valueFn func(a ...interface{}) string) {
	field := v.LookupPath(cue.ParsePath(path))
	if field.Err() != nil {
		return
	}
	if s, err := field.String(); err == nil {
		padding := 14 - len(path) - 1
		if padding < 1 {
			padding = 1
		}
		fmt.Fprintf(w, "%s%s%s\n", labelFn(path+":"), strings.Repeat(" ", padding), valueFn(fmt.Sprintf("%q", s)))
	}
}
