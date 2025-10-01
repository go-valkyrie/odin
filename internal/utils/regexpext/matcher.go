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

package regexpext

import "regexp"

type Result struct {
	matches     []string
	namedGroups map[string]int
}

func (r *Result) All() []string {
	return r.matches
}

func (r *Result) Get(pos int) string {
	return r.matches[pos]
}

func (r *Result) Length() int {
	return len(r.matches)
}

func (r *Result) Named(name string) string {
	return r.matches[r.namedGroups[name]]
}

type Matcher struct {
	expression  *regexp.Regexp
	namedGroups map[string]int
}

func NewMatcher(expression string) (*Matcher, error) {
	re, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	groups := make(map[string]int)
	for i, name := range re.SubexpNames() {
		if name != "" {
			groups[name] = i
		}
	}

	return &Matcher{expression: re, namedGroups: groups}, nil
}

func (m *Matcher) Match(text string) *Result {
	match := m.expression.FindStringSubmatch(text)

	return &Result{
		matches:     match,
		namedGroups: m.namedGroups,
	}
}
