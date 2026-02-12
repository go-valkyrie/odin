// SPDX-License-Identifier: MIT

// Package odintest provides reusable testing infrastructure for odin.
//
// # Overview
//
// This package exports utilities for testing odin bundles and commands using
// the testscript framework. It provides:
//   - DefaultParams() - Returns pre-configured testscript.Params
//   - OdinSetupCmd() - Custom command for writing odin.toml with test registries
//   - TemplateCmd() - Custom command for running template operations
//   - SetupRegistry() - In-process CUE module registry for testing
//
// # Making 'odin' Available in Tests
//
// To make the 'odin' command available in test scripts (needed for commands like
// 'odin cue mod tidy'), use testscript.Main in your TestMain function:
//
//	import (
//	    "os"
//	    "testing"
//
//	    "github.com/rogpeppe/go-internal/testscript"
//	    odincmd "go-valkyrie.com/odin/cmd/odin/cmd"
//	    internalcmd "go-valkyrie.com/odin/internal/cmd"
//	    "go-valkyrie.com/odin/pkg/odintest"
//	)
//
//	func TestMain(m *testing.M) {
//	    testscript.Main(m, map[string]func(){
//	        "odin": func() {
//	            // Set RunningEmbedded = false for proper CWD handling
//	            internalcmd.RunningEmbedded = false
//	            os.Exit(odincmd.Main())
//	        },
//	    })
//	}
//
//	func TestMyBundle(t *testing.T) {
//	    // Setup test registry
//	    host, modules, cleanup, err := odintest.SetupRegistry([]string{"./path/to/module"})
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    t.Cleanup(cleanup) // Important: use t.Cleanup, not defer!
//
//	    params := odintest.DefaultParams(
//	        odintest.WithDir("testdata"),
//	        odintest.WithCmds(map[string]func(*testscript.TestScript, bool, []string){
//	            "odin-setup": odintest.OdinSetupCmd(host, modules),
//	            "template":   odintest.TemplateCmd(ctx, registries, "", nil),
//	        }),
//	    )
//	    testscript.Run(t, params)
//	}
//
// # Important: Use t.Cleanup() for Registry Cleanup
//
// When setting up test registries, use t.Cleanup(cleanup) instead of defer cleanup().
// This ensures the registry stays alive for all subtests and subprocesses:
//
//	// Good - registry stays alive for all subprocesses
//	t.Cleanup(cleanup)
//
//	// Bad - cleanup runs too early, subprocesses can't connect
//	defer cleanup()
//
// # Custom Commands
//
// The package provides custom testscript commands:
//
//   - odin-setup - Writes odin.toml with test registry configuration
//   - template - Runs template generation (more efficient than 'exec odin template')
//
// Example test script (.txtar):
//
//	# Setup test registry and run commands
//	odin-setup
//	exec odin cue mod tidy
//	template
//	cmp stdout expected.yaml
package odintest
