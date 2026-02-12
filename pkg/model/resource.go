// SPDX-License-Identifier: MIT

package model

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"fmt"
)

type Resource struct {
	owner    *Component
	selector cue.Selector
	value    cue.Value
}

func (r *Resource) Format(f fmt.State, c rune) {
	switch c {
	case 'v':
		f.Write([]byte(
			fmt.Sprintf("#Resource %v: %v", r.selector, r.value)))
	}
}

func (r *Resource) ToYAML() ([]byte, error) {
	return yaml.Encode(r.value)
}

func (r *Resource) Name() string {
	name, _ := r.value.LookupPath(cue.ParsePath("metadata.name")).String()
	return name
}

func (r *Resource) Owner() *Component {
	return r.owner
}

func (r *Resource) Selector() cue.Selector {
	return r.selector
}

func (r *Resource) Value() cue.Value {
	return r.value
}

func newResource(owner *Component, selector cue.Selector, value cue.Value) *Resource {
	return &Resource{
		owner:    owner,
		selector: selector,
		value:    value,
	}
}
