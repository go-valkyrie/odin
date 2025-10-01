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

package template

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"

	"cuelang.org/go/cue"
	"cuelang.org/go/pkg/strings"
	"go-valkyrie.com/odin/internal/utils"
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

	env := utils.CreateCueEnvironment(opts.CacheDir, opts.Registries)

	logger.Debug("using cue environment", "env", env)

	modelOpts := []model.Option{
		model.WithEnv(env),
		model.WithLogger(logger),
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
			fmt.Printf("---\n")
		}
		if err := resource.Value().Validate(cue.Concrete(true)); err != nil {
			return err
		}
		if data, err := resource.ToYAML(); err != nil {
			return err
		} else {
			fmt.Printf("# %v.%v\n", resource.Owner().Selector(), resource.Selector())
			fmt.Print(string(data))
		}
	}

	return nil
}
