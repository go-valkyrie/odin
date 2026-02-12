// SPDX-License-Identifier: MIT

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
