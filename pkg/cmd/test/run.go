// SPDX-License-Identifier: MIT

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
	"go-valkyrie.com/odin/pkg/odintest"
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
	registryHost, modules, cleanup, err := odintest.SetupRegistry(opts.ModulePaths)
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

	// Build params options
	paramsOpts := []odintest.ParamsOption{
		odintest.WithFiles(testFiles),
		odintest.WithUpdateScripts(opts.Update),
		odintest.WithCmds(map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"odin-setup": odintest.OdinSetupCmd(registryHost, modules),
			"template":   odintest.TemplateCmd(ctx, opts.Registries, opts.CacheDir, opts.Logger),
		}),
	}

	// Create testscript params
	params := odintest.DefaultParams(paramsOpts...)

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
