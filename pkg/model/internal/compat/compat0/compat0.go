// SPDX-License-Identifier: MIT

package compat0

import "go-valkyrie.com/odin/pkg/model/internal/source"

// ApplyNamespace sets namespace via Tags. All tags must be consumed by
// @tag(...) attributes or CUE evaluation fails.
func ApplyNamespace(opts *source.LoadOptions, namespace string) {
	opts.Tags = append(opts.Tags, "namespace="+namespace)
}
