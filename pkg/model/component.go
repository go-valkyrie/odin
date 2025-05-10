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
