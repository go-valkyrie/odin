/**
 * MIT License
 *
 * Copyright (c) 2025 Stefan Nuxoll
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

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
	return cmd
}
