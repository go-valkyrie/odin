// SPDX-License-Identifier: MIT

package source

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cuelang.org/go/cue"
	"go-valkyrie.com/odin/pkg/oci"
)

type ociSource struct {
	raw     string
	ref     *oci.Reference
	tempDir string
	logger  *slog.Logger
}

func newOCI(uri string, logger *slog.Logger) (Source, error) {
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

func (s *ociSource) Prepare() error {
	tempDir, err := os.MkdirTemp("", "odin-oci-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	s.tempDir = tempDir

	ctx := context.Background()
	if err := oci.Pull(ctx, s.ref, tempDir, s.logger); err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to pull OCI bundle: %w", err)
	}
	return nil
}

func (s *ociSource) String() string {
	if s.tempDir != "" {
		return s.tempDir
	}
	return s.raw
}

func (s *ociSource) Load(ctx *cue.Context, opts *LoadOptions) (cue.Value, error) {
	if s.tempDir == "" {
		return cue.Value{}, fmt.Errorf("OCI source not prepared (call Prepare first)")
	}
	return local(s.tempDir).Load(ctx, opts)
}

func (s *ociSource) Close() error {
	if s.tempDir != "" {
		return os.RemoveAll(s.tempDir)
	}
	return nil
}
