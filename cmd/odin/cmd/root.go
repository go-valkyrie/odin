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
	"context"
	"github.com/dpotapov/slogpfx"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/config"
	"log/slog"
	"os"
	"path/filepath"
)

type rootCmd struct {
	opts       *sharedOptions
	configPath string
	logger     *slog.Logger
	debug      bool
}

func (c *rootCmd) PersistentPreRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if c.debug {
		cmd.SilenceUsage = true
	}

	if c.opts.CacheDir == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return err
		}

		c.opts.CacheDir = filepath.Join(dir, "odin")
	}

	ctx = context.WithValue(ctx, sharedOptsCtxKey, c.opts)

	var logger *slog.Logger
	if c.logger != nil {
		logger = c.logger
	} else {
		opts := &tint.Options{}

		if c.debug {
			opts.Level = slog.LevelDebug
		} else {
			opts.Level = slog.LevelInfo
		}

		handler := tint.NewHandler(colorable.NewColorableStderr(), opts)
		handler = slogpfx.NewHandler(handler, &slogpfx.HandlerOptions{
			PrefixKeys: []string{"component"},
		})

		logger = slog.New(handler)
	}

	ctx = context.WithValue(ctx, loggerCtxKey, logger)

	configManager, err := config.NewManager(logger, c.opts.ConfigPath)
	if err != nil {
		return err
	}

	if err := configManager.Load(); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, configManagerCtxKey, configManager)

	cmd.SetContext(ctx)

	return nil
}

func newRootCmd(logger *slog.Logger) *cobra.Command {
	root := &rootCmd{
		opts:   &sharedOptions{},
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:               "odin",
		Short:             "Odin CLI",
		Long:              `odin is a CLI for generating kubernetes manifests from CUE configurations`,
		PersistentPreRunE: root.PersistentPreRunE,
		SilenceErrors:     true,
	}

	cmd.PersistentFlags().BoolVarP(&root.debug,
		"debug",
		"",
		false,
		"enable debug logging")

	cmd.PersistentFlags().BoolVarP(&root.opts.Verbose,
		"verbose",
		"v",
		false,
		"enable verbose output")

	cmd.AddCommand(newCueCmd())
	cmd.AddCommand(newCacheCmd())
	cmd.AddCommand(newComponentsCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newDocsCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newPullCmd())
	cmd.AddCommand(newPushCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newTemplateCmd())
	cmd.AddCommand(newTestCmd())

	return cmd
}
