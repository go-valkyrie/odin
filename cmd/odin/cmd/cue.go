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
	"fmt"
	"github.com/spf13/cobra"
	"go-valkyrie.com/odin/internal/utils"
	"os"
)
import cuecmd "cuelang.org/go/cmd/cue/cmd"

func cuePreRunE(cmd *cobra.Command, args []string) error {
	config := configFromCommand(cmd)
	sharedOpts := sharedOptsFromCommand(cmd)

	if registries, err := config.ModuleRegistries(); err != nil {
		return err
	} else {
		registryConfig := utils.FormatRegistryConfig(registries)
		if err := os.Setenv("CUE_REGISTRY", registryConfig); err != nil {
			return err
		}
	}

	if sharedOpts.CacheDir != "" {
		if err := os.Setenv("CUE_CACHE", sharedOpts.CacheDir); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("cache directory must be set")
	}

	return nil
}

func newCueCmd() *cobra.Command {
	cmd, _ := cuecmd.New([]string{})
	cobra.EnableTraverseRunHooks = true
	cmd.PersistentPreRunE = cuePreRunE

	return cmd.Command
}
