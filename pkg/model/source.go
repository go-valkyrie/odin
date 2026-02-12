// SPDX-License-Identifier: MIT

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
