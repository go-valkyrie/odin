// SPDX-License-Identifier: MIT

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
