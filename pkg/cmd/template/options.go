// SPDX-License-Identifier: MIT

package template

import (
	"io"
	"log/slog"
)

type Options struct {
	BundlePath      string
	CacheDir        string
	Logger          *slog.Logger
	Registries      map[string]string
	ValuesLocations []string
	ValuesPath      string
	ValuesFormat    string
	Output          io.Writer
	Namespace       string
}

func DefaultOptions() *Options {
	return &Options{
		Registries:      make(map[string]string),
		Logger:          slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		ValuesLocations: []string{},
	}
}
