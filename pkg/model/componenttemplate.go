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
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/modfile"
	"cuelang.org/go/mod/module"
	"go-valkyrie.com/odin/pkg/schema"
)

type ComponentTemplate struct {
	Package string
	Name    string
	Module  string
	Version string
	Value   cue.Value
}

// ConfigSchema returns the schema fields for this template's config section.
// Options can be provided to control behavior (e.g., schema.WithExpand).
func (t *ComponentTemplate) ConfigSchema(opts ...schema.WalkOption) []*schema.SchemaField {
	configValue := t.Value.LookupPath(cue.ParsePath("config"))
	if configValue.Err() != nil {
		return nil
	}
	return schema.WalkSchema(configValue, opts...)
}

func (b *Bundle) ComponentTemplates(ctx context.Context) iter.Seq2[*ComponentTemplate, error] {
	return func(yield func(*ComponentTemplate, error) bool) {
		logger := b.logger

		logger.Debug("starting component template discovery", "sourcePath", b.sourcePath)

		// Find the module root by walking up from sourcePath
		moduleRoot, err := findModuleRoot(b.sourcePath)
		if err != nil {
			logger.Debug("failed to find module root", "err", err)
			if !yield(nil, fmt.Errorf("finding module root: %w", err)) {
				return
			}
			return
		}
		logger.Debug("found module root", "root", moduleRoot)

		// Load the module file directly from cue.mod/module.cue
		moduleFilePath := filepath.Join(moduleRoot, "cue.mod", "module.cue")
		moduleFileData, err := os.ReadFile(moduleFilePath)
		if err != nil {
			logger.Debug("failed to read module file", "err", err)
			if !yield(nil, fmt.Errorf("reading module file: %w", err)) {
				return
			}
			return
		}

		moduleFile, err := modfile.Parse(moduleFileData, moduleFilePath)
		if err != nil {
			logger.Debug("failed to parse module file", "err", err)
			if !yield(nil, fmt.Errorf("parsing module file: %w", err)) {
				return
			}
			return
		}

		logger.Debug("loaded bundle module file", "module", moduleFile.Module, "depCount", len(moduleFile.Deps))

		// Load #ComponentBase from the odin API.
		apiInsts := load.Instances([]string{"go-valkyrie.com/odin/api/v1alpha1"}, &load.Config{
			Dir: b.sourcePath,
			Env: b.env,
		})
		if len(apiInsts) == 0 {
			logger.Debug("no API instances returned")
			return
		}
		if apiInsts[0].Err != nil {
			logger.Debug("API instance has error", "err", apiInsts[0].Err)
			return
		}
		apiValue := b.ctx.BuildInstance(apiInsts[0])
		componentBase := apiValue.LookupPath(cue.ParsePath("#ComponentBase"))
		if componentBase.Err() != nil {
			logger.Debug("failed to lookup #ComponentBase", "err", componentBase.Err())
			if !yield(nil, componentBase.Err()) {
				return
			}
			return
		}

		logger.Debug("loaded #ComponentBase schema")

		// Create a module registry to fetch dependency sources.
		registry, err := modconfig.NewRegistry(&modconfig.Config{
			Env: b.env,
		})
		if err != nil {
			logger.Debug("failed to create module registry", "err", err)
			if !yield(nil, fmt.Errorf("creating module registry: %w", err)) {
				return
			}
			return
		}

		for depPath, dep := range moduleFile.Deps {
			logger.Debug("processing dependency", "dep", depPath, "version", dep.Version)

			// Skip the odin API module itself.
			if strings.HasPrefix(depPath, "go-valkyrie.com/odin/api") {
				logger.Debug("skipping odin API dependency", "dep", depPath)
				continue
			}

			// Create a module.Version for this dependency.
			modVer, err := module.NewVersion(depPath, dep.Version)
			if err != nil {
				logger.Debug("failed to create module version", "dep", depPath, "err", err)
				continue
			}

			// Fetch the module source to get its filesystem location.
			sourceLoc, err := registry.Fetch(ctx, modVer)
			if err != nil {
				logger.Debug("failed to fetch module source", "dep", depPath, "err", err)
				continue
			}

			// Extract the OS filesystem path from the SourceLoc.
			osRootFS, ok := sourceLoc.FS.(module.OSRootFS)
			if !ok {
				logger.Debug("module FS does not implement OSRootFS, skipping", "dep", depPath)
				continue
			}

			moduleDir := filepath.Join(osRootFS.OSRoot(), sourceLoc.Dir)
			logger.Debug("discovered module directory", "dep", depPath, "dir", moduleDir)

			// Use ./... wildcard from the module's directory to discover all packages.
			pkgInsts := load.Instances([]string{"./..."}, &load.Config{
				Dir: moduleDir,
				Env: b.env,
			})

			logger.Debug("discovered packages in module", "dep", depPath, "packageCount", len(pkgInsts))

			for _, inst := range pkgInsts {
				if !b.scanPackageForTemplates(inst, componentBase, depPath, dep.Version, yield) {
					return
				}
			}
		}

		// Scan local module for templates
		logger.Debug("scanning local module for templates", "moduleRoot", moduleRoot)
		localInsts := load.Instances([]string{"./..."}, &load.Config{
			Dir: moduleRoot,
			Env: b.env,
		})
		logger.Debug("discovered local packages", "packageCount", len(localInsts))
		for _, inst := range localInsts {
			if !b.scanPackageForTemplates(inst, componentBase, moduleFile.Module, "", yield) {
				return
			}
		}
	}
}

// scanPackageForTemplates scans a single package instance for component templates.
// Returns false if the caller should stop yielding (early termination requested).
func (b *Bundle) scanPackageForTemplates(
	inst *build.Instance,
	componentBase cue.Value,
	modulePath string,
	version string,
	yield func(*ComponentTemplate, error) bool,
) bool {
	logger := b.logger

	if inst.Err != nil {
		logger.Debug("skipping package with load error", "pkg", inst.ImportPath, "err", inst.Err)
		return true
	}

	logger.Debug("building package", "pkg", inst.ImportPath)

	value := b.ctx.BuildInstance(inst)
	if value.Err() != nil {
		logger.Debug("skipping package that failed to build", "pkg", inst.ImportPath, "err", value.Err())
		return true
	}

	fieldIter, err := value.Fields(cue.Definitions(true))
	if err != nil {
		logger.Debug("skipping package with no definition fields", "pkg", inst.ImportPath, "err", err)
		return true
	}

	for fieldIter.Next() {
		name := fieldIter.Selector().String()
		// Skip private definitions.
		if strings.HasPrefix(name, "_#") {
			continue
		}
		// Skip non-definition selectors.
		if !fieldIter.Selector().IsDefinition() {
			continue
		}

		logger.Debug("checking definition against #ComponentBase", "pkg", inst.ImportPath, "def", name)

		unified := fieldIter.Value().Unify(componentBase)
		if unified.Err() != nil {
			logger.Debug("definition does not unify with #ComponentBase", "pkg", inst.ImportPath, "def", name, "err", unified.Err())
			continue
		}

		// Check that apiVersion and kind are concrete after unification.
		// This filters out open types that unify with #ComponentBase
		// but aren't actually component templates.
		apiVersion := unified.LookupPath(cue.ParsePath("apiVersion"))
		kind := unified.LookupPath(cue.ParsePath("kind"))
		if apiVersion.Err() != nil || kind.Err() != nil ||
			!apiVersion.IsConcrete() || !kind.IsConcrete() {
			logger.Debug("definition unifies but lacks concrete apiVersion/kind",
				"pkg", inst.ImportPath, "def", name)
			continue
		}

		logger.Debug("found component template", "pkg", inst.ImportPath, "def", name)
		tmpl := &ComponentTemplate{
			Package: inst.ImportPath,
			Name:    name,
			Module:  modulePath,
			Version: version,
			Value:   fieldIter.Value(),
		}
		if !yield(tmpl, nil) {
			return false
		}
	}

	return true
}

// findModuleRoot walks up from the given directory to find cue.mod/module.cue.
// Returns the directory containing cue.mod, or an error if not found.
func findModuleRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	currentPath := absPath
	for {
		moduleFilePath := filepath.Join(currentPath, "cue.mod", "module.cue")
		if _, err := os.Stat(moduleFilePath); err == nil {
			return currentPath, nil
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root without finding cue.mod/module.cue
			return "", fmt.Errorf("no cue.mod/module.cue found in %s or any parent directory", startPath)
		}
		currentPath = parentPath
	}
}
