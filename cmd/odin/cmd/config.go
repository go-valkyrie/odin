// SPDX-License-Identifier: MIT

package cmd

import "github.com/spf13/cobra"

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "utilities for working with configuration",
	}

	cmd.AddCommand(newConfigEvalCmd())

	return cmd
}
