// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/config"
	"go-valkyrie.com/odin/pkg/cmd/components"
)

type componentsCmd struct {
	logger     *slog.Logger
	config     config.Manager
	cacheDir   string
	bundlePath string
	format     string
}

func (c *componentsCmd) Args(cmd *cobra.Command, args []string) error {
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

func (c *componentsCmd) PreRunE(cmd *cobra.Command, args []string) error {
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

func (c *componentsCmd) RunE(cmd *cobra.Command, args []string) error {
	opts := components.Options{
		BundlePath: c.bundlePath,
		Format:     c.format,
		CacheDir:   c.cacheDir,
		Logger:     c.logger.With("component", "components"),
	}
	globalRegistries, err := c.config.ModuleRegistries()
	if err != nil {
		return err
	}
	opts.Registries = globalRegistries
	return opts.Run(cmd.Context())
}

func newComponentsCmd() *cobra.Command {
	c := &componentsCmd{
		format: "table",
	}
	cmd := &cobra.Command{
		Use:     "components [location]",
		Short:   "list available component templates from bundle dependencies",
		Args:    c.Args,
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}

	cmd.Flags().StringVarP(&c.format, "format", "f", "table", "output format (table, json)")

	return cmd
}
