// SPDX-License-Identifier: MIT

package template

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"cuelang.org/go/cue"
	"cuelang.org/go/pkg/strings"
	"go-valkyrie.com/odin/pkg/model"
)

var (
	preserveEnvVars = []string{
		"HOME",
		"USERPROFILE",
	}
)

func (o *Options) Run(ctx context.Context) error {
	return run(ctx, *o)
}

func run(ctx context.Context, opts Options) error {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	w := opts.Output
	if w == nil {
		w = io.Writer(os.Stdout)
	}

	modelOpts := []model.Option{
		model.WithLogger(logger),
		model.WithRegistries(opts.Registries),
		model.WithCacheDir(opts.CacheDir),
	}

	if len(opts.ValuesLocations) > 0 {
		modelOpts = append(modelOpts, model.WithValues(opts.ValuesLocations))
	}

	b, err := model.LoadBundle(opts.BundlePath, modelOpts...)
	if err != nil {
		return err
	}

	if err := b.Error(); err != nil {
		return err
	}

	resources := make([]*model.Resource, 0)
	for component := range b.Components() {
		if err := component.ValidConfig(); err != nil {
			return err
		}
		resources = slices.AppendSeq(resources, component.Resources())
	}

	slices.SortFunc(resources, func(left, right *model.Resource) int {
		lname := fmt.Sprintf("%s.%s", left.Owner().Selector(), left.Selector())
		rname := fmt.Sprintf("%s.%s", right.Owner().Selector(), right.Selector())
		return strings.Compare(lname, rname)
	})

	for i, resource := range resources {
		if i > 0 {
			fmt.Fprintf(w, "---\n")
		}
		if err := resource.Value().Validate(cue.Concrete(true)); err != nil {
			return err
		}
		if data, err := resource.ToYAML(); err != nil {
			return err
		} else {
			fmt.Fprintf(w, "# %v.%v\n", resource.Owner().Selector(), resource.Selector())
			fmt.Fprint(w, string(data))
		}
	}

	return nil
}
