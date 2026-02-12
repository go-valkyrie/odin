// SPDX-License-Identifier: MIT

package cmd

import (
	"maps"
	"os"
	"strings"

	cuecmd "cuelang.org/go/cmd/cue/cmd"
	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/utils"
	"go-valkyrie.com/odin/pkg/model"
)

func cuePreRunE(cmd *cobra.Command, args []string) error {
	cfg := configFromCommand(cmd)
	sharedOpts := sharedOptsFromCommand(cmd)
	logger := loggerFromCommand(cmd)

	registries, err := cfg.ModuleRegistries()
	if err != nil {
		return err
	}

	bundleCfg, err := model.LoadConfig(".")
	if err != nil {
		return err
	}
	maps.Copy(registries, bundleCfg.Registries)

	logger.Debug("merged registries", "registries", registries)

	env := utils.CreateCueEnvironment(sharedOpts.CacheDir, registries)
	logger.Debug("using CUE environment", "env", env)

	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	return nil
}

func newCueCmd() *cobra.Command {
	cmd, _ := cuecmd.New([]string{})
	cobra.EnableTraverseRunHooks = true
	cmd.PersistentPreRunE = cuePreRunE

	return cmd.Command
}
