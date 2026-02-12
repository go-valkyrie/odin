// SPDX-License-Identifier: MIT

package integration

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	odincmd "go-valkyrie.com/odin/cmd/odin/cmd"
	internalcmd "go-valkyrie.com/odin/internal/cmd"
	"go-valkyrie.com/odin/pkg/odintest"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	// testscript.Main makes the odin command available as a real subprocess,
	// which is needed for commands like 'odin cue' that rely on CWD.
	// It calls os.Exit internally.
	testscript.Main(m, map[string]func(){
		"odin": func() {
			// Set RunningEmbedded = false so subcommands know they're running
			// in a real subprocess with proper CWD handling
			internalcmd.RunningEmbedded = false
			os.Exit(odincmd.Main())
		},
	})
}

func TestTemplateIntegration(t *testing.T) {
	ctx := context.Background()

	// Locate testdata/platform module via runtime.Caller
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	testDir := filepath.Dir(filename)
	platformPath := filepath.Join(testDir, "testdata", "platform")

	// Setup test registry with example/platform module
	host, modules, cleanup, err := odintest.SetupRegistry([]string{platformPath})
	if err != nil {
		t.Fatalf("failed to setup registry: %v", err)
	}
	t.Cleanup(cleanup) // Use t.Cleanup instead of defer so registry stays alive for all subtests/subprocesses

	// Hard-code global registries (core odin modules from ghcr.io)
	globalRegistries := map[string]string{
		"go-valkyrie.com":          "ghcr.io/go-valkyrie/cue",
		"platform.go-valkyrie.com": "ghcr.io/go-valkyrie/cue",
	}

	// Run testscript tests
	// Note: The 'odin' command is available via TestMain as a real subprocess,
	// so we don't need to wire it up via WithOdinExecutor
	params := odintest.DefaultParams(
		odintest.WithDir("testdata"),
		odintest.WithUpdateScripts(*updateGolden),
		odintest.WithCmds(map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"odin-setup": odintest.OdinSetupCmd(host, modules),
			"template":   odintest.TemplateCmd(ctx, globalRegistries, "", nil),
		}),
	)
	testscript.Run(t, params)
}
