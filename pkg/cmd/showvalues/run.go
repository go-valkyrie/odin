// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package showvalues

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"github.com/fatih/color"
	"go-valkyrie.com/odin/pkg/model"
	"go-valkyrie.com/odin/pkg/schema"
)

// Run executes the show values command.
func (o *Options) Run(ctx context.Context) error {
	// Load the bundle
	b, err := model.LoadBundle(
		o.BundlePath,
		model.WithLogger(o.Logger),
		model.WithRegistries(o.Registries),
		model.WithCacheDir(o.CacheDir),
	)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}

	// Extract values from bundle
	valuesPath := cue.ParsePath("values")
	valuesValue := b.Value().LookupPath(valuesPath)

	if !valuesValue.Exists() {
		return fmt.Errorf("bundle has no values defined")
	}

	if err := valuesValue.Err(); err != nil {
		return fmt.Errorf("bundle values has errors: %w", err)
	}

	// Determine output writer
	var w io.Writer = os.Stdout
	if o.OutputPath != "" {
		f, err := os.Create(o.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	// Format output based on requested format
	format := strings.ToLower(o.Format)
	switch format {
	case "text":
		return o.formatText(w, b, valuesValue)
	case "cue":
		return o.formatCUE(w, valuesValue)
	case "markdown", "md":
		return o.formatMarkdown(w, b, valuesValue)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, cue, markdown/md)", o.Format)
	}
}

func (o *Options) formatText(w io.Writer, b *model.Bundle, valuesValue cue.Value) error {
	// Print header with bundle name
	bold := color.New(color.Bold)
	bundleName := b.Name()
	if bundleName == "<error>" {
		bundleName = o.BundlePath
	}
	fmt.Fprintf(w, "Bundle: ")
	bold.Fprintf(w, "%s\n\n", bundleName)

	// Walk schema and format
	fields := b.ValuesSchema()
	schema.FormatSchema(w, fields, 0)

	return nil
}

func (o *Options) formatCUE(w io.Writer, valuesValue cue.Value) error {
	// Convert to syntax node with docs and optional fields
	syn := valuesValue.Syntax(
		cue.Docs(true),
		cue.Optional(true),
	)

	// Format as CUE source
	formatted, err := format.Node(syn)
	if err != nil {
		return fmt.Errorf("failed to format CUE syntax: %w", err)
	}

	_, err = w.Write(formatted)
	return err
}

func (o *Options) formatMarkdown(w io.Writer, b *model.Bundle, valuesValue cue.Value) error {
	// Print header
	bundleName := b.Name()
	if bundleName == "<error>" {
		bundleName = o.BundlePath
	}
	fmt.Fprintf(w, "# Bundle Values: %s\n\n", bundleName)

	// Walk schema and format as markdown
	fields := b.ValuesSchema()
	schema.FormatSchemaMarkdown(w, fields, 2)

	return nil
}
