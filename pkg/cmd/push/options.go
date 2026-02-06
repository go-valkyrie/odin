// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package push

import (
	"log/slog"
)

// Options holds configuration for the push command
type Options struct {
	// Reference is the OCI reference (e.g., ghcr.io/org/app:tag)
	Reference string

	// BundlePath is the path to the bundle to push
	BundlePath string

	// Logger for output
	Logger *slog.Logger
}
