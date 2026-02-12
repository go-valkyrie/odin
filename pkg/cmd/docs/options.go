// SPDX-License-Identifier: MIT

package docs

import (
	"io"
	"log/slog"
)

type Options struct {
	BundlePath string
	Reference  string
	Expand     bool
	Format     string
	OutputPath string
	NoSummary  bool
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
