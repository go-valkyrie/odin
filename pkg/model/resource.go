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
