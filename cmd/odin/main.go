// SPDX-License-Identifier: MIT

package main

import (
	"go-valkyrie.com/odin/cmd/odin/cmd"
	internalcmd "go-valkyrie.com/odin/internal/cmd"
	"os"
)

func main() {
	internalcmd.RunningEmbedded = false
	os.Exit(cmd.Main())
}
