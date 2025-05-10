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
	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/config"
	"log/slog"
)

type contextKey string

var (
	configManagerCtxKey contextKey = "configManager"
	sharedOptsCtxKey    contextKey = "sharedOpts"
	loggerCtxKey        contextKey = "logger"
)

type sharedOptions struct {
	ConfigPath string
	CacheDir   string
	Verbose    bool
}

func configFromCommand(cmd *cobra.Command) config.Manager {
	if cm, ok := cmd.Context().Value(configManagerCtxKey).(config.Manager); ok {
		return cm
	}

	return nil
}

func sharedOptsFromCommand(cmd *cobra.Command) *sharedOptions {
	if opts, ok := cmd.Context().Value(sharedOptsCtxKey).(*sharedOptions); ok {
		return opts
	}
	return nil
}

func loggerFromCommand(cmd *cobra.Command) *slog.Logger {
	if logger, ok := cmd.Context().Value(loggerCtxKey).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}
