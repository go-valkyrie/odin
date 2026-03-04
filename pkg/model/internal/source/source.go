// SPDX-License-Identifier: MIT

package source

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/load"
)

type InstanceConfiguration func(inst *build.Instance) error

type LoadOptions struct {
	Env                   []string
	Tags                  []string
	TagVars               map[string]load.TagVar
	InstanceConfiguration InstanceConfiguration
}

// Source is a source of CUE values for a bundle or values overlay.
type Source interface {
	Load(ctx *cue.Context, opts *LoadOptions) (cue.Value, error)
	fmt.Stringer
}

// New returns a Source for the given location. OCI URIs (oci://) return an
// ociSource; everything else is treated as a local filesystem path.
func New(location string, logger *slog.Logger) (Source, error) {
	if strings.HasPrefix(location, "oci://") {
		if logger == nil {
			logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		}
		return newOCI(location, logger)
	}
	return local(location), nil
}

type local string

func (s local) String() string {
	return string(s)
}

func (s local) Load(ctx *cue.Context, opts *LoadOptions) (cue.Value, error) {
	if _, err := os.Stat(string(s)); err != nil {
		return cue.Value{}, err
	}

	inst := load.Instances([]string{"."}, &load.Config{
		Dir:       string(s),
		DataFiles: true,
		Env:       opts.Env,
		Tags:      opts.Tags,
		TagVars:   opts.TagVars,
	})[0]

	if configure := opts.InstanceConfiguration; configure != nil {
		if err := configure(inst); err != nil {
			return cue.Value{}, err
		}
	}

	return ctx.BuildInstance(inst), nil
}
