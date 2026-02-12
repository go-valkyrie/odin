// SPDX-License-Identifier: MIT

package config

import (
	_ "embed"
	"go-valkyrie.com/cueconfig"
	"go-valkyrie.com/cueconfig/schema"
	"go-valkyrie.com/cueconfig/source"
)

//go:embed schema.cue
var schemaFile []byte

var configSchema *cueconfig.Schema

func init() {
	schemaSource := source.Bytes(schemaFile)
	s, err := schema.New(schemaSource)
	if err != nil {
		panic(err)
	}
	configSchema = s
}
