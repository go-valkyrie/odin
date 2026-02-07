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

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/config"
	"go-valkyrie.com/odin/pkg/cmd/test"
)

type testCmd struct {
	logger      *slog.Logger
	config      config.Manager
	cacheDir    string
	modulePaths []string
	update      bool
	testPaths   []string
	verbose     bool
}

func (c *testCmd) Args(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("requires at least 1 argument (test paths)")
	}
	c.testPaths = args
	return nil
}

func (c *testCmd) PreRunE(cmd *cobra.Command, args []string) error {
	c.logger = loggerFromCommand(cmd)
	c.config = configFromCommand(cmd)
	sharedOpts := sharedOptsFromCommand(cmd)
	c.cacheDir = sharedOpts.CacheDir
	c.verbose = sharedOpts.Verbose

	if err := ensureCacheDir(c.cacheDir); err != nil {
		return err
	}

	if len(c.modulePaths) == 0 {
		return fmt.Errorf("at least one module path (-m) is required")
	}

	return nil
}

func (c *testCmd) RunE(cmd *cobra.Command, args []string) error {
	// Load global registries from config (includes hard-coded odin registries)
	registries, err := c.config.ModuleRegistries()
	if err != nil {
		return fmt.Errorf("failed to load registries from config: %w", err)
	}

	opts := test.Options{
		ModulePaths: c.modulePaths,
		TestPaths:   c.testPaths,
		Update:      c.update,
		Verbose:     c.verbose,
		CacheDir:    c.cacheDir,
		Logger:      c.logger,
		Registries:  registries,
	}

	return opts.Run(cmd.Context())
}

func newTestCmd() *cobra.Command {
	c := &testCmd{}

	cmd := &cobra.Command{
		Use:     "test [flags] <test-paths...>",
		Short:   "Run testscript-based tests for CUE modules",
		Long:    `Run testscript-based txtar tests with an in-process CUE module registry.`,
		Args:    c.Args,
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}

	cmd.Flags().StringSliceVarP(&c.modulePaths, "module", "m", nil, "path to local CUE module to serve (required, repeatable)")
	cmd.Flags().BoolVarP(&c.update, "update", "u", false, "update golden files in txtar scripts")

	return cmd
}
