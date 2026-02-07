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
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
	"go-valkyrie.com/odin/pkg/cmd/template"
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

	// Validate module paths
	for _, mp := range opts.ModulePaths {
		moduleFile := filepath.Join(mp, "cue.mod", "module.cue")
		if _, err := os.Stat(moduleFile); err != nil {
			return fmt.Errorf("module path %s is not a valid CUE module (missing cue.mod/module.cue): %w", mp, err)
		}
	}

	// Setup in-process registry
	registryHost, modules, cleanup, err := setupRegistry(opts.ModulePaths)
	if err != nil {
		return fmt.Errorf("failed to setup registry: %w", err)
	}
	defer cleanup()

	logger.Debug("started test registry", "host", registryHost, "modules", len(modules))
	for _, mod := range modules {
		logger.Debug("serving module", "path", mod.Path)
	}

	// Discover test files
	testFiles, err := discoverTestFiles(opts.TestPaths)
	if err != nil {
		return fmt.Errorf("failed to discover test files: %w", err)
	}

	if len(testFiles) == 0 {
		return fmt.Errorf("no test files found")
	}

	logger.Info("discovered test files", "count", len(testFiles))

	// Get the odin binary path (for PATH injection)
	odinBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get odin binary path: %w", err)
	}
	odinDir := filepath.Dir(odinBinary)

	// Create testscript params
	params := testscript.Params{
		Dir:           "",
		Files:         testFiles,
		UpdateScripts: opts.Update,
		Setup: func(env *testscript.Env) error {
			// Add odin binary dir to PATH
			oldPath := env.Getenv("PATH")
			env.Setenv("PATH", odinDir+string(os.PathListSeparator)+oldPath)

			// Preserve HOME
			if home := os.Getenv("HOME"); home != "" {
				env.Setenv("HOME", home)
			}
			if userprofile := os.Getenv("USERPROFILE"); userprofile != "" {
				env.Setenv("USERPROFILE", userprofile)
			}

			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"odin-setup": func(ts *testscript.TestScript, neg bool, args []string) {
				if neg {
					ts.Fatalf("odin-setup does not support negation")
				}
				if len(args) > 0 {
					ts.Fatalf("odin-setup takes no arguments")
				}

				// Write odin.toml with registry entries
				odinToml := createOdinToml(registryHost, modules)
				ts.Check(os.WriteFile(ts.MkAbs("odin.toml"), []byte(odinToml), 0644))
			},
			"template": func(ts *testscript.TestScript, neg bool, args []string) {
				if neg {
					ts.Fatalf("template does not support negation")
				}

				// Parse arguments (bundle path and optional flags)
				bundlePath := "."
				var valuesFiles []string

				for i := 0; i < len(args); i++ {
					arg := args[i]
					if arg == "-f" || arg == "--values" {
						if i+1 >= len(args) {
							ts.Fatalf("flag %s requires an argument", arg)
						}
						valuesFiles = append(valuesFiles, ts.MkAbs(args[i+1]))
						i++
					} else {
						bundlePath = arg
					}
				}

				// Merge registries: start with global registries (includes hard-coded odin registries)
				// then overlay with local bundle config registries
				allRegistries := make(map[string]string)
				for k, v := range opts.Registries {
					allRegistries[k] = v
				}

				bundleConfig, err := model.LoadConfig(".")
				if err != nil {
					ts.Fatalf("failed to load config: %v", err)
				}
				for k, v := range bundleConfig.Registries {
					allRegistries[k] = v
				}

				// Create template options
				var output strings.Builder
				templateOpts := template.Options{
					BundlePath:      ts.MkAbs(bundlePath),
					CacheDir:        opts.CacheDir,
					Logger:          opts.Logger,
					Registries:      allRegistries,
					ValuesLocations: valuesFiles,
					Output:          &output,
				}

				// Run template command
				if err := templateOpts.Run(ctx); err != nil {
					ts.Fatalf("template failed: %v", err)
				}

				// Write output to stdout
				ts.Stdout().Write([]byte(output.String()))
			},
		},
	}

	// Create a custom test runner
	runner := &runner{
		logger:  logger,
		verbose: opts.Verbose,
		passed:  0,
		failed:  0,
	}

	// Run tests
	testscript.RunT(&runT{runner: runner}, params)

	// Print summary
	total := runner.passed + runner.failed
	logger.Info("test summary", "total", total, "passed", runner.passed, "failed", runner.failed)

	if runner.failed > 0 {
		return fmt.Errorf("%d test(s) failed", runner.failed)
	}

	return nil
}

// discoverTestFiles finds all .txtar files in the given paths
func discoverTestFiles(paths []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			// Search for *.txtar files in directory (non-recursive)
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if strings.HasSuffix(entry.Name(), ".txtar") {
					fullPath := filepath.Join(path, entry.Name())
					absPath, err := filepath.Abs(fullPath)
					if err != nil {
						return nil, err
					}
					if !seen[absPath] {
						files = append(files, absPath)
						seen[absPath] = true
					}
				}
			}
		} else {
			// Single file
			if !strings.HasSuffix(path, ".txtar") {
				return nil, fmt.Errorf("test file must have .txtar extension: %s", path)
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			if !seen[absPath] {
				files = append(files, absPath)
				seen[absPath] = true
			}
		}
	}

	return files, nil
}

// createOdinToml generates odin.toml content with registry entries (only for test modules)
func createOdinToml(registryHost string, modules []moduleInfo) string {
	var sb strings.Builder

	for _, mod := range modules {
		sb.WriteString(fmt.Sprintf("[[registries]]\n"))
		sb.WriteString(fmt.Sprintf("module-prefix = \"%s\"\n", mod.Path))
		sb.WriteString(fmt.Sprintf("registry = \"%s\"\n", registryHost))
		sb.WriteString("\n")
	}

	return sb.String()
}

// runT implements testscript.T interface
type runT struct {
	runner *runner
}

type runner struct {
	logger  *slog.Logger
	verbose bool
	passed  int
	failed  int
}

var (
	skipPanic = "skip"
	failPanic = "fail"
)

func (t *runT) Run(name string, f func(t testscript.T)) {
	defer func() {
		if r := recover(); r != nil {
			// Check if it's a skip or fail panic
			if r == skipPanic {
				return
			}
			if r == failPanic {
				t.runner.failed++
				t.runner.logger.Error("test failed", "name", name)
				return
			}
			// Re-panic if it's something else
			panic(r)
		}
		t.runner.passed++
		if t.runner.verbose {
			t.runner.logger.Info("test passed", "name", name)
		}
	}()

	ts := &testScriptT{
		name:    name,
		runner:  t.runner,
		verbose: t.runner.verbose,
	}
	f(ts)
}

func (t *runT) FailNow() {
	panic(failPanic)
}

func (t *runT) Fatal(args ...interface{}) {
	fmt.Println(args...)
	panic(failPanic)
}

func (t *runT) Skip(args ...interface{}) {
	fmt.Println(args...)
	panic(skipPanic)
}

func (t *runT) Log(args ...interface{}) {
	if t.runner.verbose {
		fmt.Println(args...)
	}
}

func (t *runT) Parallel() {
	// No-op for sequential execution
}

func (t *runT) Verbose() bool {
	return t.runner.verbose
}

// testScriptT implements testscript.T for individual tests
type testScriptT struct {
	name    string
	runner  *runner
	verbose bool
}

func (t *testScriptT) Skip(args ...interface{}) {
	t.Log(args...)
	panic(skipPanic)
}

func (t *testScriptT) Fatal(args ...interface{}) {
	t.Log(args...)
	panic(failPanic)
}

func (t *testScriptT) Parallel() {
	// No-op for sequential execution
}

func (t *testScriptT) Log(args ...interface{}) {
	if t.verbose {
		t.runner.logger.Info(fmt.Sprint(args...), "test", t.name)
	}
}

func (t *testScriptT) FailNow() {
	panic(failPanic)
}

func (t *testScriptT) Run(name string, f func(t testscript.T)) {
	// Nested runs not supported in our simple implementation
	f(t)
}

func (t *testScriptT) Verbose() bool {
	return t.verbose
}
