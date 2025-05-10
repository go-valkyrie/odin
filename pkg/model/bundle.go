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

package model

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
	"fmt"
	"go-valkyrie.com/odin/internal/schema"
	"io"
	"iter"
	"log/slog"
	"os"
)

func configureValuesInstance(inst *build.Instance) error {
	if inst.BuildFiles != nil {
		return nil
	}

	for _, f := range inst.OrphanedFiles {
		reader, err := os.Open(f.Filename)
		if err != nil {
			continue
		}
		defer reader.Close()

		switch f.Encoding {
		case "yaml":
			fallthrough
		case "json":
			if file, err := yaml.Extract(f.Filename, reader); err != nil {
				continue
			} else if err := inst.AddSyntax(file); err != nil {
				return err
			}
		default:
			continue
		}

	}

	return nil
}

type Option func(bundle *bundleLoader) error

type bundleLoader struct {
	ctx          *cue.Context
	env          []string
	logger       *slog.Logger
	source       modelSource
	valuesSource modelSource
}

func WithContext(ctx *cue.Context) Option {
	return func(l *bundleLoader) error {
		l.ctx = ctx
		return nil
	}
}

func WithEnv(env []string) Option {
	return func(l *bundleLoader) error {
		l.env = env
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(l *bundleLoader) error {
		l.logger = logger
		return nil
	}
}

func WithValues(path string) Option {
	return func(l *bundleLoader) error {
		l.valuesSource = localSource(path)
		return nil
	}
}

func dumpCue(value cue.Value) {
	fmt.Printf("%v\n", value)
}

func (l *bundleLoader) Load() (*Bundle, error) {
	if l.source == nil {
		return nil, fmt.Errorf("modelSource is required")
	}
	logger := l.logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	b := &Bundle{
		ctx: l.ctx,
		env: l.env,
	}

	if b.ctx == nil {
		b.ctx = cuecontext.New()
	}

	if b.env == nil {
		b.env = os.Environ()
	}

	logger.Debug("loading bundle", "source", l.source.String())

	if value, err := l.source.Load(b.ctx, &sourceLoadOptions{
		Env: b.env,
	}); err != nil {
		return nil, err
	} else {
		b.value = value
	}

	//if logger.Enabled(nil, slog.LevelDebug) {
	//	bundleInstance := b.value.BuildInstance()
	//	bundleFiles := make([]string, 0, len(bundleInstance.BuildFiles))
	//	for _, file := range bundleInstance.BuildFiles {
	//		bundleFiles = append(bundleFiles, file.Filename)
	//	}
	//	logger.Debug("loaded bundle", "name", b.Name(), "files", bundleFiles)
	//}

	if bundleSchema, err := schema.LoadBundleSchema(b.ctx); err != nil {
		return nil, err
	} else {
		b.value = b.value.Unify(bundleSchema)
	}

	logger.Debug("loading values", "source", l.valuesSource.String())
	if l.valuesSource != nil {
		if _b, err := b.LoadValues(l.valuesSource); err != nil {
			return nil, err
		} else {
			b = _b
		}
	}

	return b, nil
}

func LoadBundle(bundlePath string, options ...Option) (*Bundle, error) {
	l := &bundleLoader{}

	if source, err := newSource(bundlePath); err != nil {
		return nil, err
	} else {
		l.source = source
	}

	for _, option := range options {
		if err := option(l); err != nil {
			return nil, err
		}
	}

	return l.Load()
}

type Bundle struct {
	ctx   *cue.Context
	env   []string
	value cue.Value
}

func (b *Bundle) GoString() string {
	return fmt.Sprintf("#Bundle & %v", b.value)
}

func (b *Bundle) LoadValues(source modelSource) (*Bundle, error) {
	values, err := source.Load(b.ctx, &sourceLoadOptions{
		Env:                   b.env,
		InstanceConfiguration: configureValuesInstance,
	})
	if err != nil {
		return nil, err
	}

	value := b.value.FillPath(cue.ParsePath("values"), values)

	newBundle := &Bundle{
		ctx:   b.ctx,
		value: value,
	}
	return newBundle, nil
}

func (b *Bundle) Components() iter.Seq[*Component] {
	return func(yield func(*Component) bool) {
		i, err := b.value.LookupPath(cue.ParsePath("components")).Fields(cue.Definitions(false))
		if err != nil {
			return
		}

		for i.Next() {
			yield(newComponent(i.Selector(), i.Value()))
		}
	}
}

func (b *Bundle) Error() error {
	return b.value.Err()
}

func (b *Bundle) Name() string {
	if name, err := b.value.LookupPath(cue.ParsePath("metadata.name")).String(); err != nil {
		return "<error>"
	} else {
		return name
	}

}

func (b *Bundle) Value() cue.Value {
	return b.value
}
