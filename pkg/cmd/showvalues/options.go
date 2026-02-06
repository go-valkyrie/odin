// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package showvalues

import (
	"log/slog"
)

// Options contains the configuration for showing bundle values.
type Options struct {
	// BundlePath is the path to the bundle.
	BundlePath string

	// Format is the output format (text, cue, markdown).
	Format string

	// OutputPath is the file to write output to (empty for stdout).
	OutputPath string

	// CacheDir is the cache directory for bundle loading.
	CacheDir string

	// Logger is the logger to use.
	Logger *slog.Logger

	// Registries maps module prefixes to OCI registries.
	Registries map[string]string
}
