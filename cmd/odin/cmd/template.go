// SPDX-License-Identifier: MIT

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
	namespace   string
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

func (c *templateCmd) RunE(cmd *cobra.Command, args []string) error {
	opts := template.Options{
		BundlePath:      c.bundlePath,
		CacheDir:        c.cacheDir,
		Logger:          c.logger.With("component", "template"),
		ValuesLocations: c.valuesFiles,
		Namespace:       c.namespace,
	}
	// Load global registries first
	globalRegistries, err := c.config.ModuleRegistries()
	if err != nil {
		return err
	}
	// Pass global registries; bundle-local registries will be merged inside the model loader.
	opts.Registries = globalRegistries
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
	cmd.Flags().StringVar(&c.namespace, "namespace", "", "Namespace to use for @tag(namespace) in CUE")

	return cmd
}
