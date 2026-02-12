// SPDX-License-Identifier: MIT

package schema

import (
	"cuelang.org/go/cue"
	_ "embed"
)

//go:embed bundle.cue
var bundleSchema []byte

func LoadBundleSchema(ctx *cue.Context) (cue.Value, error) {
	schema := ctx.CompileBytes(bundleSchema, cue.Filename("bundle.cue"))
	if schema.Err() != nil {
		return cue.Value{}, schema.Err()
	}
	return schema.LookupPath(cue.ParsePath("#Bundle")), nil
}
