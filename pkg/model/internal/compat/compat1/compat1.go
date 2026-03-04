// SPDX-License-Identifier: MIT

package compat1

import (
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/load"
	"go-valkyrie.com/odin/pkg/model/internal/source"
)

// ApplyNamespace sets namespace via TagVars so unreferenced vars are silently
// ignored. @tag attributes must use the var=<name> form:
// @tag(namespace, var=namespace).
func ApplyNamespace(opts *source.LoadOptions, namespace string) {
	ns := namespace
	opts.TagVars = map[string]load.TagVar{
		"namespace": {
			Func: func() (ast.Expr, error) {
				return ast.NewString(ns), nil
			},
		},
	}
}
