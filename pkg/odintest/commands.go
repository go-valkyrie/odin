// SPDX-License-Identifier: MIT

package odintest

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
	"go-valkyrie.com/odin/pkg/cmd/template"
	"go-valkyrie.com/odin/pkg/model"
)

// OdinSetupCmd returns a testscript command function that writes odin.toml with registry entries.
// This command must be called before running 'exec cue mod tidy' in test scripts.
func OdinSetupCmd(registryHost string, modules []ModuleInfo) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("odin-setup does not support negation")
		}
		if len(args) > 0 {
			ts.Fatalf("odin-setup takes no arguments")
		}

		// Write odin.toml with registry entries
		odinToml := CreateOdinToml(registryHost, modules)
		ts.Check(os.WriteFile(ts.MkAbs("odin.toml"), []byte(odinToml), 0644))
	}
}

// TemplateCmd returns a testscript command function that runs the odin template command.
// It merges global registries with local bundle config registries, allowing test modules
// to override real modules while preserving access to core odin modules.
//
// Supports negation (! prefix) for expected failures.
// Supports -f/--values flags for values overlays.
func TemplateCmd(ctx context.Context, globalRegistries map[string]string, cacheDir string, logger *slog.Logger) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		// Parse arguments (bundle path and optional flags)
		bundlePath := "."
		var valuesFiles []string
		var namespace string

		for i := 0; i < len(args); i++ {
			arg := args[i]
			if arg == "-f" || arg == "--values" {
				if i+1 >= len(args) {
					ts.Fatalf("flag %s requires an argument", arg)
				}
				valuesFiles = append(valuesFiles, ts.MkAbs(args[i+1]))
				i++
			} else if arg == "--namespace" || arg == "-n" {
				if i+1 >= len(args) {
					ts.Fatalf("flag %s requires an argument", arg)
				}
				namespace = args[i+1]
				i++
			} else {
				bundlePath = arg
			}
		}

		// Merge registries: start with global registries (includes hard-coded odin registries)
		// then overlay with local bundle config registries
		allRegistries := make(map[string]string)
		for k, v := range globalRegistries {
			allRegistries[k] = v
		}

		bundleConfig, err := model.LoadConfig(".")
		if err != nil {
			if !neg {
				ts.Fatalf("failed to load config: %v", err)
			}
			return
		}
		for k, v := range bundleConfig.Registries {
			allRegistries[k] = v
		}

		// Create template options
		var output strings.Builder
		templateOpts := template.Options{
			BundlePath:      ts.MkAbs(bundlePath),
			CacheDir:        cacheDir,
			Logger:          logger,
			Registries:      allRegistries,
			ValuesLocations: valuesFiles,
			Namespace:       namespace,
			Output:          &output,
		}

		// Run template command
		err = templateOpts.Run(ctx)
		if neg {
			if err == nil {
				ts.Fatalf("template succeeded, but expected failure")
			}
			return
		}
		if err != nil {
			ts.Fatalf("template failed: %v", err)
		}

		// Write output to stdout
		ts.Stdout().Write([]byte(output.String()))
	}
}
