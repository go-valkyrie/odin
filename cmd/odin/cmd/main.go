// SPDX-License-Identifier: MIT

package cmd

import (
	cueerrors "cuelang.org/go/cue/errors"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log/slog"
	"os"
)

type Command struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	root   *cobra.Command
	logger *slog.Logger
}

func (c *Command) Execute(args []string) error {
	cmd := newRootCmd(c.logger)
	cmd.SetArgs(args)

	cmd.SetIn(c.In)

	cmd.SetOut(c.Out)

	cmd.SetErr(c.Err)

	if err := cmd.Execute(); err == nil {
		return nil
	} else if e := cueerrors.Error(nil); errors.As(err, &e) {
		cmd.PrintErrln("Error processing CUE files:")
		errs := cueerrors.Errors(e)
		for _, err := range errs {
			cmd.PrintErrln(err.Error())
			if pos := err.InputPositions(); len(pos) > 0 {
				cmd.PrintErrln(" Positions:")
				for _, pos := range pos {
					cmd.PrintErrln("  ", pos.String())
				}
			} else {
				cmd.PrintErrln(" Position:", err.Position().String())
			}
		}
		return err
	} else {
		cmd.PrintErr(fmt.Sprintf("Error: %v\n", err))
		return err
	}
}

type Options struct {
	Logger *slog.Logger
}

func New(opts *Options) *Command {
	command := &Command{
		logger: opts.Logger,
	}

	return command
}

func Main() int {
	opts := &Options{}

	command := New(opts)

	if err := command.Execute(os.Args[1:]); err != nil {
		return 1
	}

	return 0
}
