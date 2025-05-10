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
