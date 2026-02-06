// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/pkg/cmd/push"
)

type pushCmd struct {
	reference  string
	bundlePath string
}

func newPushCmd() *cobra.Command {
	p := &pushCmd{}

	cmd := &cobra.Command{
		Use:   "push <oci-reference> [bundle-path]",
		Short: "Push a bundle to an OCI registry",
		Long: `Push a bundle to an OCI registry as a gzip-compressed tarball.

The reference should be in the format: registry/repository:tag or oci://registry/repository:tag

Examples:
  odin push ghcr.io/org/app:v1
  odin push ghcr.io/org/app:v1 ./my-bundle
  odin push oci://registry.example.com/project/bundle:latest`,
		Args: cobra.RangeArgs(1, 2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			p.reference = args[0]

			// Handle bundle path
			if len(args) > 1 {
				p.bundlePath = args[1]
			} else {
				// Default to current directory, but find bundle root
				root, err := findBundleRoot(".")
				if err != nil {
					return fmt.Errorf("no bundle found in current directory (use explicit path or run from bundle directory): %w", err)
				}
				p.bundlePath = root
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := loggerFromCommand(cmd)

			opts := push.Options{
				Reference:  p.reference,
				BundlePath: p.bundlePath,
				Logger:     logger,
			}

			return push.Run(cmd.Context(), opts)
		},
	}

	return cmd
}
