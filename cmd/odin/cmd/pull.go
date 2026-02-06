// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/pkg/cmd/pull"
)

type pullCmd struct {
	reference string
	outputDir string
}

func newPullCmd() *cobra.Command {
	p := &pullCmd{}

	cmd := &cobra.Command{
		Use:   "pull <oci-reference>",
		Short: "Pull a bundle from an OCI registry",
		Long: `Pull a bundle from an OCI registry and extract it to a local directory.

The reference should be in the format: registry/repository:tag or oci://registry/repository:tag

If no output directory is specified, defaults to {bundle-name}-{tag} in the current directory.

Examples:
  odin pull ghcr.io/org/app:v1
  odin pull ghcr.io/org/app:v1 -o ./my-bundle
  odin pull oci://registry.example.com/project/bundle:latest -o /tmp/bundle`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			p.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := loggerFromCommand(cmd)

			opts := pull.Options{
				Reference: p.reference,
				OutputDir: p.outputDir,
				Logger:    logger,
			}

			return pull.Run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&p.outputDir, "output", "o", "", "output directory (default: {bundle-name}-{tag})")

	return cmd
}
