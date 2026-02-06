// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package push

import (
	"context"
	"fmt"

	"go-valkyrie.com/odin/pkg/oci"
)

// Run executes the push command
func Run(ctx context.Context, opts Options) error {
	// Parse OCI reference
	ref, err := oci.ParseReference(opts.Reference)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}

	// Push bundle
	if err := oci.Push(ctx, ref, opts.BundlePath, opts.Logger); err != nil {
		return fmt.Errorf("failed to push bundle: %w", err)
	}

	return nil
}
