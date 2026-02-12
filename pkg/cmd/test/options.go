// SPDX-License-Identifier: MIT

package test

import (
	"io"
	"log/slog"
)

type Options struct {
	ModulePaths   []string          // local CUE modules to serve
	TestPaths     []string          // txtar files or directories
	Update        bool              // -u flag
	Verbose       bool
	CacheDir      string
	Logger        *slog.Logger
	Registries    map[string]string // global registries (includes hard-coded odin registries)
}

func DefaultOptions() *Options {
	return &Options{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}
}
