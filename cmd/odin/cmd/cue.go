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
