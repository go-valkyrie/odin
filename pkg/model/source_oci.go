// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cuelang.org/go/cue"
	"go-valkyrie.com/odin/pkg/oci"
)

// ociSource implements modelSource for OCI registry bundles
type ociSource struct {
	raw     string         // original oci:// URI
	ref     *oci.Reference // parsed OCI reference
	tempDir string         // populated after Prepare()
	logger  *slog.Logger
}

// newOCISource creates a new OCI source
func newOCISource(uri string, logger *slog.Logger) (modelSource, error) {
	ref, err := oci.ParseReference(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid OCI reference: %w", err)
	}

	return &ociSource{
		raw:    uri,
		ref:    ref,
		logger: logger,
	}, nil
}

// Prepare pulls the bundle from OCI registry to a temp directory
func (s *ociSource) Prepare() error {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "odin-oci-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	s.tempDir = tempDir

	// Pull bundle
	ctx := context.Background()
	if err := oci.Pull(ctx, s.ref, tempDir, s.logger); err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to pull OCI bundle: %w", err)
	}

	return nil
}

// String returns the path to the extracted bundle after Prepare()
func (s *ociSource) String() string {
	if s.tempDir != "" {
		return s.tempDir
	}
	return s.raw
}

// Load delegates to localSource after Prepare() has been called
func (s *ociSource) Load(ctx *cue.Context, opts *sourceLoadOptions) (cue.Value, error) {
	if s.tempDir == "" {
		return cue.Value{}, fmt.Errorf("OCI source not prepared (call Prepare first)")
	}

	// Delegate to local source
	local := localSource(s.tempDir)
	return local.Load(ctx, opts)
}

// Close cleans up the temp directory
func (s *ociSource) Close() error {
	if s.tempDir != "" {
		return os.RemoveAll(s.tempDir)
	}
	return nil
}
