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
	"fmt"
	"io"
	"iter"
	"log/slog"
	"maps"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
	"go-valkyrie.com/odin/internal/schema"
	"go-valkyrie.com/odin/internal/utils"
	pkgschema "go-valkyrie.com/odin/pkg/schema"
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
	registries   map[string]string
	cacheDir     string
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

// WithRegistries provides a default set of CUE registries, these may be overridden by the bundle config.
func WithRegistries(reg map[string]string) Option {
	return func(l *bundleLoader) error {
		l.registries = reg
		return nil
	}
}

// WithCacheDir provides the cache directory so the loader can build the CUE env.
func WithCacheDir(cacheDir string) Option {
	return func(l *bundleLoader) error {
		l.cacheDir = cacheDir
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(l *bundleLoader) error {
		l.logger = logger
		return nil
	}
}

func WithValues(locations []string) Option {
	return func(l *bundleLoader) error {
		if source, err := newValuesSource(locations); err != nil {
			return err
		} else {
			l.valuesSource = source

			return nil
		}
	}
}

func (l *bundleLoader) Load() (*Bundle, error) {
	if l.source == nil {
		return nil, fmt.Errorf("modelSource is required")
	}

	logger := l.logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	if l.registries == nil {
		l.registries = map[string]string{}
	}

	// Check if source needs preparation (e.g., OCI sources need to pull first)
	type preparableSource interface {
		Prepare() error
	}
	if p, ok := l.source.(preparableSource); ok {
		if err := p.Prepare(); err != nil {
			return nil, fmt.Errorf("failed to prepare source: %w", err)
		}
	}

	b, err := newBundle(l.ctx)
	if err != nil {
		return nil, err
	}

	bundlePath := l.source.String()
	b.sourcePath = bundlePath
	b.logger = logger
	cfg, err := LoadConfig(bundlePath)
	if err != nil {
		return nil, err
	}

	b.addRegistries(l.registries)
	b.addRegistries(cfg.Registries)

	b.env = utils.CreateCueEnvironment(l.cacheDir, b.Registries())

	logger.Debug("using CUE environment", "env", b.env)

	logger.Debug("loading bundle", "source", l.source.String())

	if value, err := l.source.Load(b.ctx, &sourceLoadOptions{
		Env: b.env,
	}); err != nil {
		return nil, err
	} else {
		b.value = value
	}

	if bundleSchema, err := schema.LoadBundleSchema(b.ctx); err != nil {
		return nil, err
	} else {
		b.value = b.value.Unify(bundleSchema)
	}

	if l.valuesSource != nil {
		logger.Debug("loading values", "source", l.valuesSource.String())
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

	// Apply options first so we have logger if needed
	for _, option := range options {
		if err := option(l); err != nil {
			return nil, err
		}
	}

	// Create source with logger
	if source, err := newSource(bundlePath, l.logger); err != nil {
		return nil, err
	} else {
		l.source = source
	}

	return l.Load()
}

type Bundle struct {
	ctx        *cue.Context
	env        []string
	value      cue.Value
	registries map[string]string
	sourcePath string
	logger     *slog.Logger
}

func newBundle(cuectx *cue.Context) (*Bundle, error) {
	if cuectx == nil {
		cuectx = cuecontext.New()
	}

	b := &Bundle{
		ctx:        cuectx,
		env:        make([]string, 0, 4),
		registries: make(map[string]string),
	}

	return b, nil
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
		ctx:        b.ctx,
		env:        b.env,
		value:      value,
		registries: b.registries,
		sourcePath: b.sourcePath,
		logger:     b.logger,
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

func (b *Bundle) addRegistries(registries map[string]string) {
	if b.registries != nil {
		maps.Copy(b.registries, registries)
	}
}

func (b *Bundle) Registries() map[string]string {
	return b.registries
}

// ValuesSchema returns the schema fields for the bundle's values section,
// with validation pattern constraints filtered out.
func (b *Bundle) ValuesSchema() []*pkgschema.SchemaField {
	valuesValue := b.value.LookupPath(cue.ParsePath("values"))
	if !valuesValue.Exists() || valuesValue.Err() != nil {
		return nil
	}
	fields := pkgschema.WalkSchema(valuesValue)
	filterValuesSchemaPatterns(fields)
	return fields
}

// filterValuesSchemaPatterns removes the [string]: {...} pattern constraint
// from the components field in bundle values. This pattern is just validation
// scaffolding in the bundle schema and not meaningful documentation for users.
func filterValuesSchemaPatterns(fields []*pkgschema.SchemaField) {
	for _, field := range fields {
		if field.Name == "components" && len(field.Children) > 0 {
			// Filter out empty pattern constraints from components
			filtered := make([]*pkgschema.SchemaField, 0, len(field.Children))
			for _, child := range field.Children {
				// Skip [string]: {...} pattern with no actual fields
				if child.IsPattern && len(child.Children) == 0 && child.Type == "{...}" {
					continue
				}
				filtered = append(filtered, child)
			}
			field.Children = filtered
		}
	}
}
