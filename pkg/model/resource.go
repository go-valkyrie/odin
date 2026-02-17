// SPDX-License-Identifier: MIT

package model

import (
	"bytes"
	"fmt"

	"cuelang.org/go/cue"
	"gopkg.in/yaml.v3"
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
	var resourceMap map[string]interface{}
	if err := r.value.Decode(&resourceMap); err != nil {
		return nil, err
	}

	// Encode with 2-space indentation
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(resourceMap); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
