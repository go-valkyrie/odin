// SPDX-License-Identifier: MIT

package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string // Returns temp dir path
		wantErr       bool
		wantRegisties map[string]string
		wantCompat    int
	}{
		{
			name: "valid odin.toml",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				content := `[[registries]]
module-prefix = "example.com/module"
registry = "registry.example.com"

[[registries]]
module-prefix = "other.com/foo"
registry = "registry.other.com"
`
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantRegisties: map[string]string{
				"example.com/module": "registry.example.com",
				"other.com/foo":      "registry.other.com",
			},
		},
		{
			name: "missing odin.toml - returns empty config",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			wantRegisties: map[string]string{},
		},
		{
			name: "empty odin.toml",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(""), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantRegisties: map[string]string{},
		},
		{
			name: "entries with empty fields are skipped",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				content := `[[registries]]
module-prefix = "example.com/module"
registry = "registry.example.com"

[[registries]]
module-prefix = ""
registry = "registry.empty.com"

[[registries]]
module-prefix = "other.com/foo"
registry = ""
`
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantRegisties: map[string]string{
				"example.com/module": "registry.example.com",
			},
		},
		{
			name: "invalid toml",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				content := `[[registries]
invalid toml syntax
`
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantErr: true,
		},
		{
			name: "empty bundle path defaults to current directory",
			setupFunc: func(t *testing.T) string {
				// Change to temp dir and create odin.toml there
				dir := t.TempDir()
				origDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get working directory: %v", err)
				}
				t.Cleanup(func() {
					os.Chdir(origDir)
				})
				if err := os.Chdir(dir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}
				content := `[[registries]]
module-prefix = "test.com/module"
registry = "registry.test.com"
`
				if err := os.WriteFile("odin.toml", []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return "" // Empty string to test default behavior
			},
			wantRegisties: map[string]string{
				"test.com/module": "registry.test.com",
			},
		},
		{
			name: "compat = 1 parses correctly",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				content := `compat = 1
`
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantRegisties: map[string]string{},
			wantCompat:    1,
		},
		{
			name: "missing compat defaults to 0",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				content := `[[registries]]
module-prefix = "example.com/module"
registry = "registry.example.com"
`
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantRegisties: map[string]string{
				"example.com/module": "registry.example.com",
			},
			wantCompat: 0,
		},
		{
			name: "compat = 0 explicitly set",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				content := `compat = 0
`
				if err := os.WriteFile(filepath.Join(dir, "odin.toml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				return dir
			},
			wantRegisties: map[string]string{},
			wantCompat:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundlePath := tt.setupFunc(t)
			cfg, err := LoadConfig(bundlePath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(cfg.Registries) != len(tt.wantRegisties) {
				t.Errorf("got %d registries, want %d", len(cfg.Registries), len(tt.wantRegisties))
			}
			for k, v := range tt.wantRegisties {
				if got := cfg.Registries[k]; got != v {
					t.Errorf("Registries[%q] = %q, want %q", k, got, v)
				}
			}
			if cfg.Compat != tt.wantCompat {
				t.Errorf("Compat = %d, want %d", cfg.Compat, tt.wantCompat)
			}
		})
	}
}
