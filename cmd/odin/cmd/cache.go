// SPDX-License-Identifier: MIT

package cmd

import "github.com/spf13/cobra"

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "utilities for working with odin's cache",
	}

	cmd.AddCommand(newCacheCleanCmd())

	return cmd
}
