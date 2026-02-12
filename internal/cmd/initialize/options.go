// SPDX-License-Identifier: MIT

package initialize

import (
	"io"
	"log/slog"
)

type Options struct {
	BundlePath string
	BundleName string
	ModulePath string
	Logger     *slog.Logger
	Prompt     bool
}

func NewOptions() *Options {
	return &Options{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}
}
