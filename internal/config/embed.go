// SPDX-License-Identifier: MIT

package config

import (
	_ "embed"
)

//go:embed template.cue
var userConfigTemplate []byte
