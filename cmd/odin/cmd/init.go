// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	internalcmd "go-valkyrie.com/odin/internal/cmd"
	"go-valkyrie.com/odin/internal/cmd/initialize"
)

type initCmd struct {
	opts       *initialize.Options
	sharedOpts *sharedOptions
}

func (c *initCmd) Args(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}

	return nil
}

func (c *initCmd) PreRunE(cmd *cobra.Command, args []string) error {
	if internalcmd.RunningEmbedded {
		return fmt.Errorf("'odin init' cannot be run when used as a library, see https://github.com/cue-lang/cue/issues/3916'")
	}
	if path := cmd.Flags().Arg(0); path != "" {
		c.opts.BundlePath = path
	} else {
		c.opts.BundlePath = "."
	}

	logger := loggerFromCommand(cmd).With("component", "init")

	c.opts.Logger = logger

	return nil
}

func (c *initCmd) RunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	return c.opts.Run(ctx)
}

func newInitCmd() *cobra.Command {
	c := &initCmd{
		opts: initialize.NewOptions(),
	}

	cmd := &cobra.Command{
		Use:     "init [path]",
		Short:   "initialize a new odin bundle",
		Args:    c.Args,
		PreRunE: c.PreRunE,
		RunE:    c.RunE,
	}
	cmd.Flags().BoolVarP(&c.opts.Prompt, "prompt", "p", true, "use interactive prompts to configure values")
	cmd.Flags().StringVarP(&c.opts.ModulePath, "module", "m", "", "specify name of generated cue module (infers from git remote by default)")
	cmd.Flags().IntVar(&c.opts.Compat, "compat", 1, "compat level to write into odin.toml (0=legacy Tags, 1=TagVars)")
	return cmd
}
