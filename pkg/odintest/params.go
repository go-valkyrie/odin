// SPDX-License-Identifier: MIT

package odintest

import (
	"os"

	"github.com/rogpeppe/go-internal/testscript"
)

// DefaultParams returns a testscript.Params with sensible defaults for testing odin.
// The returned params can be customized using functional options or by directly
// modifying the returned struct before passing it to testscript.Run.
//
// Note: To make the 'odin' command available in test scripts, use testscript.Main
// in your TestMain function. See the package documentation for examples.
func DefaultParams(opts ...ParamsOption) testscript.Params {
	params := testscript.Params{
		Setup: func(env *testscript.Env) error {
			// Preserve HOME and USERPROFILE for CUE module cache
			if home := os.Getenv("HOME"); home != "" {
				env.Setenv("HOME", home)
			}
			if userprofile := os.Getenv("USERPROFILE"); userprofile != "" {
				env.Setenv("USERPROFILE", userprofile)
			}
			return nil
		},
		Cmds: map[string]func(*testscript.TestScript, bool, []string){},
	}

	for _, opt := range opts {
		opt(&params)
	}

	return params
}

// ParamsOption is a functional option for customizing testscript.Params.
type ParamsOption func(*testscript.Params)

// WithUpdateScripts sets the UpdateScripts flag (equivalent to -update flag).
func WithUpdateScripts(update bool) ParamsOption {
	return func(p *testscript.Params) {
		p.UpdateScripts = update
	}
}

// WithDir sets the directory containing test scripts.
func WithDir(dir string) ParamsOption {
	return func(p *testscript.Params) {
		p.Dir = dir
	}
}

// WithFiles sets the list of test script files to run.
func WithFiles(files []string) ParamsOption {
	return func(p *testscript.Params) {
		p.Files = files
	}
}

// WithCmds adds additional custom commands to the params.
func WithCmds(cmds map[string]func(*testscript.TestScript, bool, []string)) ParamsOption {
	return func(p *testscript.Params) {
		if p.Cmds == nil {
			p.Cmds = make(map[string]func(*testscript.TestScript, bool, []string))
		}
		for k, v := range cmds {
			p.Cmds[k] = v
		}
	}
}
