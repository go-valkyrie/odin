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

package test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/mod/modfile"
	"cuelang.org/go/mod/modregistrytest"
)

type moduleInfo struct {
	Path string // e.g. "platform.vituity.com/common"
}

// setupRegistry starts an in-process CUE module registry serving all local modules at v0.0.0-test
func setupRegistry(modulePaths []string) (host string, modules []moduleInfo, cleanup func(), err error) {
	if len(modulePaths) == 0 {
		return "", nil, nil, fmt.Errorf("no module paths provided")
	}

	tempDir, err := os.MkdirTemp("", "odin-test-registry-*")
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanupTemp := func() {
		os.RemoveAll(tempDir)
	}

	modules = make([]moduleInfo, 0, len(modulePaths))

	for _, modulePath := range modulePaths {
		// Read module.cue to get module path
		moduleFilePath := filepath.Join(modulePath, "cue.mod", "module.cue")
		data, err := os.ReadFile(moduleFilePath)
		if err != nil {
			cleanupTemp()
			return "", nil, nil, fmt.Errorf("failed to read %s: %w", moduleFilePath, err)
		}

		mf, err := modfile.Parse(data, moduleFilePath)
		if err != nil {
			cleanupTemp()
			return "", nil, nil, fmt.Errorf("failed to parse %s: %w", moduleFilePath, err)
		}

		if mf.Module == "" {
			cleanupTemp()
			return "", nil, nil, fmt.Errorf("module path empty in %s", moduleFilePath)
		}

		// Copy module to temp dir with modregistrytest naming convention
		// module/path@v0.0.0-test becomes module_path_v0.0.0-test
		version := "v0.0.0-test"
		registryName := strings.ReplaceAll(mf.Module, "/", "_") + "_" + version
		destPath := filepath.Join(tempDir, registryName)

		if err := copyDir(modulePath, destPath); err != nil {
			cleanupTemp()
			return "", nil, nil, fmt.Errorf("failed to copy module %s: %w", modulePath, err)
		}

		modules = append(modules, moduleInfo{
			Path: mf.Module,
		})
	}

	// Start the registry
	registry, err := modregistrytest.New(os.DirFS(tempDir), "")
	if err != nil {
		cleanupTemp()
		return "", nil, nil, fmt.Errorf("failed to start registry: %w", err)
	}

	cleanup = func() {
		registry.Close()
		cleanupTemp()
	}

	host = registry.Host()
	return
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, 0644)
	})
}
