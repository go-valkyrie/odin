// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package pull

import (
	"context"
	"fmt"
	"path/filepath"

	"go-valkyrie.com/odin/pkg/oci"
)

// Run executes the pull command
func Run(ctx context.Context, opts Options) error {
	// Parse OCI reference
	ref, err := oci.ParseReference(opts.Reference)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}

	// Determine output directory if not specified
	outputDir := opts.OutputDir
	if outputDir == "" {
		// Generate default: {lastComponent}-{tag}
		outputDir = fmt.Sprintf("%s-%s", ref.LastComponent(), ref.Reference)
		// Make it absolute from current directory
		outputDir, err = filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("failed to resolve output directory: %w", err)
		}
	}

	// Pull bundle
	if err := oci.Pull(ctx, ref, outputDir, opts.Logger); err != nil {
		return fmt.Errorf("failed to pull bundle: %w", err)
	}

	opts.Logger.Info("bundle extracted", "directory", outputDir)
	return nil
}
