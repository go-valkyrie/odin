// SPDX-License-Identifier: MIT

package config

import (
	"go-valkyrie.com/cueconfig/source"
)

func loadSource(path string) (source.Source, error) {
	if path == "" {
		// Use UserConfigSource for global config
		return source.UserConfig(&source.UserConfigOptions{
			Vendor:   "Valkyrie",
			App:      "odin",
			Filename: "config.cue",
			Template: map[string][]byte{
				"config.cue": userConfigTemplate,
			},
		})
	} else {
		return source.Path(path)
	}
}
