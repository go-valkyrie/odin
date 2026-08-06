// SPDX-License-Identifier: MIT

package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// testConfigureValuesInstance mirrors pkg/model.configureValuesInstance
// (unexported in a different package) so orphaned yaml/json data files get
// parsed into the instance, matching the real Values.Load call path.
func testConfigureValuesInstance(inst *build.Instance) error {
	if inst.BuildFiles != nil {
		return nil
	}
	for _, f := range inst.OrphanedFiles {
		reader, err := os.Open(f.Filename)
		if err != nil {
			continue
		}
		defer reader.Close()

		switch f.Encoding {
		case "yaml", "json":
			if file, err := yaml.Extract(f.Filename, reader); err != nil {
				continue
			} else if err := inst.AddSyntax(file); err != nil {
				return err
			}
		}
	}
	return nil
}

// TestValuesLoad_NestedModuleRegression pins the CUE v0.17 "directory is in a
// nested module" regression: a values directory living inside the bundle's
// own module, loaded while the process cwd sits under an unrelated ancestor
// module. Without ModuleRoot anchored to the bundle, CUE resolves its module
// root by walking up from cwd, lands on the wrong (ancestor) module, and then
// flags the bundle's own cue.mod as "nested" beneath it.
func TestValuesLoad_NestedModuleRegression(t *testing.T) {
	root := t.TempDir()
	outer := filepath.Join(root, "outer")
	bundle := filepath.Join(outer, "bundle")

	writeFile(t, filepath.Join(outer, "cue.mod", "module.cue"), `module: "outer.example.com@v0"
language: version: "v0.8.0"
`)
	writeFile(t, filepath.Join(bundle, "cue.mod", "module.cue"), `module: "bundle.example.com@v0"
language: version: "v0.8.0"
`)
	writeFile(t, filepath.Join(bundle, "values", "values.yaml"), "foo: bar\n")

	t.Chdir(outer)

	values, err := NewValues([]string{"./bundle/values"})
	if err != nil {
		t.Fatalf("NewValues: %v", err)
	}

	t.Run("without ModuleRoot reproduces the bug", func(t *testing.T) {
		ctx := cuecontext.New()
		v, err := values.Load(ctx, &LoadOptions{InstanceConfiguration: testConfigureValuesInstance})
		if err != nil {
			t.Fatalf("Load returned Go error: %v", err)
		}
		if v.Err() == nil {
			t.Fatal("expected embedded error for nested module, got nil")
		}
		if !strings.Contains(v.Err().Error(), "nested module") {
			t.Fatalf("expected 'nested module' error, got: %v", v.Err())
		}
	})

	t.Run("with ModuleRoot pinned to the bundle, it works", func(t *testing.T) {
		ctx := cuecontext.New()
		v, err := values.Load(ctx, &LoadOptions{
			ModuleRoot:            bundle,
			InstanceConfiguration: testConfigureValuesInstance,
		})
		if err != nil {
			t.Fatalf("Load returned Go error: %v", err)
		}
		if v.Err() != nil {
			t.Fatalf("unexpected error: %v", v.Err())
		}
		foo, err := v.LookupPath(cue.ParsePath("foo")).String()
		if err != nil {
			t.Fatalf("looking up foo: %v", err)
		}
		if foo != "bar" {
			t.Fatalf("foo = %q, want %q", foo, "bar")
		}
	})
}

func TestNewValues_FormatParsing(t *testing.T) {
	dir := t.TempDir()
	plainFile := filepath.Join(dir, "values.cue")
	writeFile(t, plainFile, "foo: \"bar\"\n")

	tests := []struct {
		name       string
		location   string
		wantFormat string
		wantPath   string
	}{
		{"untagged", plainFile, "", plainFile},
		{"yaml", "yaml: " + plainFile, "yaml", plainFile},
		{"json", "json: " + plainFile, "json", plainFile},
		{"cue (directory, since cue: locations must be dirs)", "cue: " + dir, "cue", dir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := NewValues([]string{tt.location})
			if err != nil {
				t.Fatalf("NewValues(%q): %v", tt.location, err)
			}
			if len(values.locations) != 1 {
				t.Fatalf("got %d locations, want 1", len(values.locations))
			}
			if got := values.locations[0].format; got != tt.wantFormat {
				t.Fatalf("format = %q, want %q", got, tt.wantFormat)
			}
			if got := values.locations[0].path; got != tt.wantPath {
				t.Fatalf("path = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func TestNewValues_CueLocationMustBeDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "values.cue")
	writeFile(t, file, "foo: \"bar\"\n")

	if _, err := NewValues([]string{"cue: " + file}); err == nil {
		t.Fatal("expected error for cue: location that is a file, got nil")
	} else if !strings.Contains(err.Error(), "must be a directory") {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := NewValues([]string{"cue: " + dir}); err != nil {
		t.Fatalf("expected no error for cue: location that is a directory, got: %v", err)
	}
}

// TestValuesLoad_CueGroupUnification proves a cue:-tagged location is loaded
// as its own independent, self-rooted module (decoupled from the bundle and
// from any default-group location) and that its result is unified with the
// default group's result.
func TestValuesLoad_CueGroupUnification(t *testing.T) {
	bundleDir := t.TempDir()
	writeFile(t, filepath.Join(bundleDir, "cue.mod", "module.cue"), `module: "bundle.example.com@v0"
language: version: "v0.8.0"
`)
	writeFile(t, filepath.Join(bundleDir, "default-values.yaml"), "fromDefault: 1\n")

	externalDir := t.TempDir()
	writeFile(t, filepath.Join(externalDir, "cue.mod", "module.cue"), `module: "external.example.com@v0"
language: version: "v0.8.0"
`)
	writeFile(t, filepath.Join(externalDir, "values.cue"), `package external

fromCue: "external"
`)

	values, err := NewValues([]string{
		filepath.Join(bundleDir, "default-values.yaml"),
		"cue: " + externalDir,
	})
	if err != nil {
		t.Fatalf("NewValues: %v", err)
	}

	ctx := cuecontext.New()
	v, err := values.Load(ctx, &LoadOptions{
		ModuleRoot:            bundleDir,
		InstanceConfiguration: testConfigureValuesInstance,
	})
	if err != nil {
		t.Fatalf("Load returned Go error: %v", err)
	}
	if v.Err() != nil {
		t.Fatalf("unexpected error: %v", v.Err())
	}

	fromDefault, err := v.LookupPath(cue.ParsePath("fromDefault")).Int64()
	if err != nil {
		t.Fatalf("looking up fromDefault: %v", err)
	}
	if fromDefault != 1 {
		t.Fatalf("fromDefault = %d, want 1", fromDefault)
	}

	fromCue, err := v.LookupPath(cue.ParsePath("fromCue")).String()
	if err != nil {
		t.Fatalf("looking up fromCue: %v", err)
	}
	if fromCue != "external" {
		t.Fatalf("fromCue = %q, want %q", fromCue, "external")
	}
}
