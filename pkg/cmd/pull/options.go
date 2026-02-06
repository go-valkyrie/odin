// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package pull

import (
	"log/slog"
)

// Options holds configuration for the pull command
type Options struct {
	// Reference is the OCI reference (e.g., ghcr.io/org/app:tag)
	Reference string

	// OutputDir is the directory to extract the bundle to
	OutputDir string

	// Logger for output
	Logger *slog.Logger
}
