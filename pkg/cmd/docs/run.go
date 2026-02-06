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
	"strings"

	"cuelang.org/go/cue"
	"go-valkyrie.com/odin/pkg/docs"
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

	tmpl, err := docs.ResolveReference(opts.Reference, templates)
	if err != nil {
		return err
	}

	// Print header
	fmt.Printf("%s %s\n", tmpl.Package, tmpl.Name)
	fmt.Println()

	// Print doc comments
	docComments := tmpl.Value.Doc()
	for _, cg := range docComments {
		text := strings.TrimSpace(cg.Text())
		if text != "" {
			fmt.Println(text)
			fmt.Println()
		}
	}

	// Print apiVersion and kind if available
	printConcreteField(tmpl.Value, "apiVersion")
	printConcreteField(tmpl.Value, "kind")

	// Print config schema
	configValue := tmpl.Value.LookupPath(cue.ParsePath("config"))
	if configValue.Err() == nil {
		fmt.Println()
		fmt.Println("Config:")
		fields := docs.WalkSchema(configValue, opts.Expand)
		docs.FormatSchema(os.Stdout, fields, 2)
	}

	return nil
}

func printConcreteField(v cue.Value, path string) {
	field := v.LookupPath(cue.ParsePath(path))
	if field.Err() != nil {
		return
	}
	if s, err := field.String(); err == nil {
		fmt.Printf("%-14s%q\n", path+":", s)
	}
}
