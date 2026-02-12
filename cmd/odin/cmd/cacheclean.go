// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io/fs"
	"log/slog"
)

type cacheCleanCmd struct {
	logger     *slog.Logger
	sharedOpts *sharedOptions
}

func (c *cacheCleanCmd) PreRunE(cmd *cobra.Command, args []string) error {
	c.logger = loggerFromCommand(cmd)
	c.sharedOpts = sharedOptsFromCommand(cmd)

	return nil
}

func (c *cacheCleanCmd) RunE(cmd *cobra.Command, args []string) error {
	cacheDir := c.sharedOpts.CacheDir
	verbose := c.sharedOpts.Verbose
	fmt.Printf("cleaning cache directory %s\n", cacheDir)
	dirFS := afero.NewBasePathFs(afero.NewOsFs(), cacheDir).(*afero.BasePathFs)
	if err := fs.WalkDir(afero.NewIOFS(dirFS), ".", func(path string, d fs.DirEntry, err error) error {
		realpath, _ := dirFS.RealPath(path)
		if verbose {
			fmt.Printf("chmod 755 %s\n", realpath)
		}
		if err := dirFS.Chmod(path, 0755); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if verbose {
			fmt.Printf("rm %s\n", realpath)
		}
		return dirFS.Remove(path)
	}); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("rm -r %s\n", cacheDir)
	}

	return dirFS.RemoveAll(".")
}

func newCacheCleanCmd() *cobra.Command {
	c := &cacheCleanCmd{}

	cmd := &cobra.Command{
		Use:     "clean",
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}

	return cmd
}
