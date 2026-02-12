// SPDX-License-Identifier: MIT

package components

import (
	"io"
	"log/slog"
)

type Options struct {
	BundlePath string
	Format     string
	CacheDir   string
	Logger     *slog.Logger
	Registries map[string]string
}

func DefaultOptions() *Options {
	return &Options{
		Registries: make(map[string]string),
		Logger:     slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}
}
