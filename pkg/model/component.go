// SPDX-License-Identifier: MIT

package model

import (
	"cuelang.org/go/cue"
	"fmt"
	"iter"
)

type Component struct {
	selector cue.Selector
	value    cue.Value
}

func (c *Component) GoString() string {
	return fmt.Sprintf("#Component{selector: %s, value: %v}", c.selector, c.value)
}

func (c *Component) Selector() cue.Selector {
	return c.selector
}

func (c *Component) Resources() iter.Seq[*Resource] {
	return func(yield func(*Resource) bool) {
		i, err := c.value.LookupPath(cue.ParsePath("resources")).Fields(cue.Definitions(false))
		if err != nil {
			return
		}

		for i.Next() {
			yield(newResource(c, i.Selector(), i.Value()))
		}
	}
}

func (c *Component) ValidConfig() error {
	return c.value.LookupPath(cue.ParsePath("config")).Validate(cue.Final())
}

func newComponent(selector cue.Selector, value cue.Value) *Component {
	return &Component{
		selector: selector,
		value:    value,
	}
}
