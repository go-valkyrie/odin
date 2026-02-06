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

package model

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/load"
)

type instanceConfiguration func(inst *build.Instance) error

type sourceLoadOptions struct {
	Env                   []string
	InstanceConfiguration instanceConfiguration
}

type modelSource interface {
	Load(ctx *cue.Context, opts *sourceLoadOptions) (cue.Value, error)
	fmt.Stringer
}

func newSource(location string, logger *slog.Logger) (modelSource, error) {
	if strings.HasPrefix(location, "oci://") {
		if logger == nil {
			logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		}
		return newOCISource(location, logger)
	}
	return localSource(location), nil
}

type localSource string

func (s localSource) String() string {
	return string(s)
}

func (s localSource) Load(ctx *cue.Context, opts *sourceLoadOptions) (cue.Value, error) {
	if _, err := os.Stat(string(s)); err != nil {
		return cue.Value{}, err
	}

	inst := load.Instances([]string{"."}, &load.Config{
		Dir:       string(s),
		DataFiles: true,
		Env:       opts.Env,
	})[0]

	if configure := opts.InstanceConfiguration; configure != nil {
		if err := configure(inst); err != nil {
			return cue.Value{}, err
		}
	}

	return ctx.BuildInstance(inst), nil
}
