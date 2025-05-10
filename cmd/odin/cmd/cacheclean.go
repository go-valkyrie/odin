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
