// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

type configEvalCmd struct {
}

func (c *configEvalCmd) RunE(cmd *cobra.Command, args []string) error {
	config := configFromCommand(cmd)
	if data, err := config.Evaluated(); err != nil {
		return err
	} else {
		fmt.Println(string(data))
	}
	return nil
}

func newConfigEvalCmd() *cobra.Command {
	c := &configEvalCmd{}

	cmd := &cobra.Command{
		Use:  "eval",
		RunE: c.RunE,
	}

	return cmd
}
