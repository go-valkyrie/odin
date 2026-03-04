// SPDX-License-Identifier: MIT

package compat

import (
	"fmt"

	"go-valkyrie.com/odin/pkg/model/internal/compat/compat0"
	"go-valkyrie.com/odin/pkg/model/internal/compat/compat1"
	"go-valkyrie.com/odin/pkg/model/internal/source"
)

// LoadInput contains the loader state passed to BuildLoadOptions.
type LoadInput struct {
	Env       []string
	Namespace string
}

// Policy holds compat-level-specific behaviour for bundle loading. Add a new
// private field when a future compat level introduces a new behavioral
// variation, and handle it in BuildLoadOptions.
type Policy struct {
	applyNamespace func(opts *source.LoadOptions, namespace string)
}

// BuildLoadOptions constructs a source.LoadOptions from the given input,
// delegating each compat-specific concern to the appropriate policy function.
func (p *Policy) BuildLoadOptions(in LoadInput) *source.LoadOptions {
	opts := &source.LoadOptions{Env: in.Env}
	if in.Namespace != "" {
		p.applyNamespace(opts, in.Namespace)
	}
	return opts
}

// NewPolicy returns the Policy for the given compat level, or an error if the
// level is not supported by this version of odin.
func NewPolicy(level int) (*Policy, error) {
	switch level {
	case 0:
		return &Policy{
			applyNamespace: compat0.ApplyNamespace,
		}, nil
	case 1:
		return &Policy{
			applyNamespace: compat1.ApplyNamespace,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported compat level %d: upgrade odin to a newer version", level)
	}
}
