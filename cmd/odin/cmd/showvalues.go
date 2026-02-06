// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/config"
	"go-valkyrie.com/odin/pkg/cmd/showvalues"
)

type showValuesCmd struct {
	logger     *slog.Logger
	config     config.Manager
	cacheDir   string
	bundlePath string
	format     string
	outputPath string
}

func (c *showValuesCmd) Args(cmd *cobra.Command, args []string) error {
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

func (c *showValuesCmd) PreRunE(cmd *cobra.Command, args []string) error {
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

func (c *showValuesCmd) RunE(cmd *cobra.Command, args []string) error {
	opts := showvalues.Options{
		BundlePath: c.bundlePath,
		Format:     c.format,
		OutputPath: c.outputPath,
		CacheDir:   c.cacheDir,
		Logger:     c.logger.With("component", "show-values"),
	}
	globalRegistries, err := c.config.ModuleRegistries()
	if err != nil {
		return err
	}
	opts.Registries = globalRegistries
	return opts.Run(cmd.Context())
}

func newShowValuesCmd() *cobra.Command {
	c := &showValuesCmd{
		format: "text",
	}
	cmd := &cobra.Command{
		Use:   "values [location]",
		Short: "Show the values schema for a bundle",
		Long: `Show the values schema for a bundle.

The values schema defines what configuration options the bundle exposes,
including types, defaults, and documentation. This is useful for understanding
what can be configured in a bundle before templating it.

Examples:
  # Show values for current bundle
  odin show values

  # Show values for bundle at path
  odin show values ./path/to/bundle

  # Output as CUE source
  odin show values -f cue

  # Output as markdown
  odin show values -f markdown -o values.md`,
		Args:    c.Args,
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}

	cmd.Flags().StringVarP(&c.format, "format", "f", "text", "Output format (text, cue, markdown/md)")
	cmd.Flags().StringVarP(&c.outputPath, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}
