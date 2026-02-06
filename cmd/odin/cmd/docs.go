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
	"go-valkyrie.com/odin/pkg/cmd/docs"
)

type docsCmd struct {
	logger     *slog.Logger
	config     config.Manager
	cacheDir   string
	bundlePath string
	reference  string
	expand     bool
	format     string
	outputPath string
	noSummary  bool
}

func (c *docsCmd) Args(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exactly one argument required: the component template reference")
	}
	c.reference = args[0]
	return nil
}

func (c *docsCmd) PreRunE(cmd *cobra.Command, args []string) error {
	sharedOpts := sharedOptsFromCommand(cmd)
	c.cacheDir = sharedOpts.CacheDir
	c.logger = loggerFromCommand(cmd)
	c.config = configFromCommand(cmd)

	if err := ensureCacheDir(c.cacheDir); err != nil {
		return err
	}

	// Auto-discover bundle root if using default path
	if c.bundlePath == "." {
		root, err := findBundleRoot(".")
		if err != nil {
			return err
		}
		c.bundlePath = root
	}

	return nil
}

func (c *docsCmd) RunE(cmd *cobra.Command, args []string) error {
	// Validate format-specific requirements
	if (c.format == "markdown-multi" || c.format == "mdm" || c.format == "mdbook" || c.format == "mdb") && c.outputPath == "" {
		return fmt.Errorf("format %q requires -o/--output to specify a directory path", c.format)
	}
	if c.noSummary && c.format != "mdbook" && c.format != "mdb" {
		return fmt.Errorf("--no-summary is only valid with mdbook format")
	}

	opts := docs.Options{
		BundlePath: c.bundlePath,
		Reference:  c.reference,
		Expand:     c.expand,
		Format:     c.format,
		OutputPath: c.outputPath,
		NoSummary:  c.noSummary,
		CacheDir:   c.cacheDir,
		Logger:     c.logger.With("component", "docs"),
	}
	globalRegistries, err := c.config.ModuleRegistries()
	if err != nil {
		return err
	}
	opts.Registries = globalRegistries
	return opts.Run(cmd.Context())
}

func newDocsCmd() *cobra.Command {
	c := &docsCmd{
		bundlePath: ".",
		format:     "text",
	}
	cmd := &cobra.Command{
		Use:   "docs <reference>",
		Short: "show documentation for a component template or package",
		Long: `Display documentation for a component template or all templates under a package path.

Reference formats:
  - Single template: "deployment", "workload.Deployment", "pkg/path:#Definition"
  - Package path: "platform.vituity.com/common", "platform.vituity.com/common/workload"

Output formats (-f/--format):
  - text (default): colored terminal output
  - markdown/md: single markdown document (concatenated if multiple templates)
  - markdown-multi/mdm: one markdown file per template (requires -o directory)
  - mdbook/mdb: same as mdm plus SUMMARY.md (requires -o directory)`,
		Args:    c.Args,
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}

	cmd.Flags().StringVarP(&c.bundlePath, "bundle", "b", ".", "bundle location")
	cmd.Flags().BoolVar(&c.expand, "expand", false, "recursively expand referenced definitions inline")
	cmd.Flags().StringVarP(&c.format, "format", "f", "text", "output format (text, markdown/md, markdown-multi/mdm, mdbook/mdb)")
	cmd.Flags().StringVarP(&c.outputPath, "output", "o", "", "output file or directory path (required for mdm/mdb formats)")
	cmd.Flags().BoolVar(&c.noSummary, "no-summary", false, "disable SUMMARY.md generation in mdbook format")

	return cmd
}
