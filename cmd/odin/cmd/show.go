// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show information about bundles",
		Long:  "Show information about bundles, such as their values schema.",
	}

	cmd.AddCommand(newShowValuesCmd())

	return cmd
}
