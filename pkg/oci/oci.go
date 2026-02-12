// SPDX-License-Identifier: MIT

package oci

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

// Reference wraps an OCI reference
type Reference struct {
	Registry   string
	Repository string
	Reference  string // tag or digest
}

// ParseReference parses an OCI reference string, optionally stripping the oci:// scheme
func ParseReference(raw string) (*Reference, error) {
	// Strip oci:// scheme if present
	raw = strings.TrimPrefix(raw, "oci://")

	// Parse using ORAS library
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid reference: must include registry and repository")
	}

	registry := parts[0]
	repoAndRef := parts[1]

	// Split repository and reference (tag or digest)
	var repository, reference string
	if idx := strings.LastIndex(repoAndRef, "@"); idx != -1 {
		// Digest reference
		repository = repoAndRef[:idx]
		reference = repoAndRef[idx+1:]
	} else if idx := strings.LastIndex(repoAndRef, ":"); idx != -1 {
		// Tag reference
		repository = repoAndRef[:idx]
		reference = repoAndRef[idx+1:]
	} else {
		// No reference, default to latest
		repository = repoAndRef
		reference = "latest"
	}

	return &Reference{
		Registry:   registry,
		Repository: repository,
		Reference:  reference,
	}, nil
}

// String returns the full reference string
func (r *Reference) String() string {
	sep := ":"
	if strings.HasPrefix(r.Reference, "sha256:") {
		sep = "@"
	}
	return fmt.Sprintf("%s/%s%s%s", r.Registry, r.Repository, sep, r.Reference)
}

// LastComponent returns the last path segment of the repository
func (r *Reference) LastComponent() string {
	parts := strings.Split(r.Repository, "/")
	return parts[len(parts)-1]
}

// podmanAuthPath resolves the Podman auth file path
func podmanAuthPath() string {
	if p := os.Getenv("REGISTRY_AUTH_FILE"); p != "" {
		return p
	}
	if runtime.GOOS == "linux" {
		if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
			return filepath.Join(xdg, "containers", "auth.json")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "containers", "auth.json")
}

// newCredentialStore creates a new credentials store for OCI auth
// Primary: Docker config ($DOCKER_CONFIG/config.json or ~/.docker/config.json)
// Respects credsStore, credHelpers, and auths entries
// Fallback: Podman auth file ($REGISTRY_AUTH_FILE or $XDG_RUNTIME_DIR/containers/auth.json)
func newCredentialStore() (*auth.Client, error) {
	dockerStore, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create docker credential store: %w", err)
	}

	// Fallback: Podman auth file
	// Podman uses $XDG_RUNTIME_DIR/containers/auth.json (Linux)
	// or ~/.config/containers/auth.json (macOS/Windows)
	store := credentials.Store(dockerStore)
	if podmanPath := podmanAuthPath(); podmanPath != "" {
		podmanStore, err := credentials.NewStore(podmanPath, credentials.StoreOptions{})
		if err == nil {
			store = credentials.NewStoreWithFallbacks(dockerStore, podmanStore)
		}
	}

	return &auth.Client{
		Credential: credentials.Credential(store),
	}, nil
}

// Push pushes a bundle to an OCI registry
func Push(ctx context.Context, ref *Reference, bundlePath string, annotations map[string]string, logger *slog.Logger) error {
	logger.Info("pushing bundle", "reference", ref.String(), "path", bundlePath)

	// Create file store from bundle directory
	fileStore, err := file.New(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer func() {
		if cerr := fileStore.Close(); cerr != nil {
			logger.Debug("failed to close file store", "error", cerr)
		}
	}()

	// TODO: Handle .odinignore - file.Store doesn't support ignore patterns directly
	// For now, we'll skip this and users should clean their bundle before pushing

	// Add the directory - this creates a tar layer with proper annotations
	layerDesc, err := fileStore.Add(ctx, ".", "", bundlePath)
	if err != nil {
		return fmt.Errorf("failed to add bundle directory: %w", err)
	}

	// Pack into a manifest with the layer
	packOpts := oras.PackManifestOptions{
		Layers: []ocispec.Descriptor{layerDesc},
	}
	if len(annotations) > 0 {
		packOpts.ManifestAnnotations = annotations
	}
	manifestDesc, err := oras.PackManifest(ctx, fileStore, oras.PackManifestVersion1_1, "application/vnd.odin.bundle.v1", packOpts)
	if err != nil {
		return fmt.Errorf("failed to pack manifest: %w", err)
	}

	// Tag the manifest
	if err := fileStore.Tag(ctx, manifestDesc, ref.Reference); err != nil {
		return fmt.Errorf("failed to tag manifest: %w", err)
	}

	// Set up remote repository
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Use plain HTTP for localhost
	if strings.HasPrefix(ref.Registry, "localhost") {
		repo.PlainHTTP = true
	}

	// Set up auth
	authClient, err := newCredentialStore()
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}
	repo.Client = authClient

	// Copy from file store to remote
	desc, err := oras.Copy(ctx, fileStore, ref.Reference, repo, ref.Reference, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("failed to push to registry: %w", err)
	}

	logger.Info("bundle pushed successfully", "digest", desc.Digest.String())
	return nil
}

// Pull pulls a bundle from an OCI registry
func Pull(ctx context.Context, ref *Reference, outputDir string, logger *slog.Logger) error {
	logger.Info("pulling bundle", "reference", ref.String(), "output", outputDir)

	// Set up remote repository
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Use plain HTTP for localhost
	if strings.HasPrefix(ref.Registry, "localhost") {
		repo.PlainHTTP = true
	}

	// Set up auth
	authClient, err := newCredentialStore()
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}
	repo.Client = authClient

	// Create file store for output directory
	fileStore, err := file.New(outputDir)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer func() {
		if cerr := fileStore.Close(); cerr != nil {
			logger.Debug("failed to close file store", "error", cerr)
		}
	}()

	// Copy from remote to file store - this automatically unpacks
	_, err = oras.Copy(ctx, repo, ref.Reference, fileStore, ref.Reference, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}

	logger.Info("bundle pulled successfully")
	return nil
}
