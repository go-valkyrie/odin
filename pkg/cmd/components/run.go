// SPDX-License-Identifier: MIT

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
