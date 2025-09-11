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
	"go-valkyrie.com/odin/pkg/cmd/template"
)

type templateCmd struct {
	logger      *slog.Logger
	config      config.Manager
	cacheDir    string
	bundlePath  string
	valuesFiles []string
}

func (c *templateCmd) Args(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}
	if len(args) > 0 {
		c.bundlePath = args[0]
	} else {
		c.bundlePath = "."
	}
	return nil
}

func (c *templateCmd) PreRunE(cmd *cobra.Command, args []string) error {
	sharedOpts := sharedOptsFromCommand(cmd)
	c.cacheDir = sharedOpts.CacheDir
	c.logger = loggerFromCommand(cmd)
	c.config = configFromCommand(cmd)

	if err := ensureCacheDir(c.cacheDir); err != nil {
		return err
	}

	return nil
}

func (c *templateCmd) RunE(cmd *cobra.Command, args []string) error {
	opts := template.Options{
		BundlePath:      c.bundlePath,
		CacheDir:        c.cacheDir,
		Logger:          c.logger.With("component", "template"),
		ValuesLocations: c.valuesFiles,
	}
	if registries, err := c.config.ModuleRegistries(); err != nil {
		return err
	} else {
		opts.Registries = registries
	}
	return opts.Run(cmd.Context())
}

func newTemplateCmd() *cobra.Command {
	c := &templateCmd{}
	cmd := &cobra.Command{
		Use:     "template [location]",
		Short:   "render templates from a bundle",
		Args:    c.Args,
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}
	cmd.Flags().StringArrayVarP(&c.valuesFiles, "values", "f", []string{}, "Values files")

	return cmd
}
