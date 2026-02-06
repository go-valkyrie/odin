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

package components

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"

	"go-valkyrie.com/odin/pkg/model"
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

	switch opts.Format {
	case "table":
		return runTable(templates)
	case "json":
		return runJSON(templates)
	default:
		return fmt.Errorf("unsupported output format: %q (supported: table, json)", opts.Format)
	}
}

func runTable(templates []*model.ComponentTemplate) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PACKAGE\tDEFINITION\tVERSION")

	for _, tmpl := range templates {
		fmt.Fprintf(w, "%s\t%s\t%s\n", tmpl.Package, tmpl.Name, tmpl.Version)
	}

	return w.Flush()
}

type componentJSON struct {
	Package string `json:"package"`
	Name    string `json:"name"`
	Module  string `json:"module"`
	Version string `json:"version"`
}

func runJSON(templates []*model.ComponentTemplate) error {
	components := make([]componentJSON, 0, len(templates))
	for _, tmpl := range templates {
		components = append(components, componentJSON{
			Package: tmpl.Package,
			Name:    tmpl.Name,
			Module:  tmpl.Module,
			Version: tmpl.Version,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(components)
}
