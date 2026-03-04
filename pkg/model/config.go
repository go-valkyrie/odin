// SPDX-License-Identifier: MIT

package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds the parsed odin.toml configuration.
//
// Compat controls tag-injection behavior:
//
//   - 0 (legacy): uses load.Config.Tags — all tags in Tags must be consumed by
//     @tag(...) attributes or CUE evaluation fails.
//   - 1: uses load.Config.TagVars — unreferenced vars are silently ignored,
//     but @tag attributes must use the var=<name> form, e.g.
//     @tag(namespace, var=namespace).
//
// See docs/COMPAT.md for migration guidance.
type Config struct {
	Registries map[string]string
	Compat     int
}

type registryEntry struct {
	ModulePrefix string `toml:"module-prefix"`
	Registry     string `toml:"registry"`
}

type tomlRoot struct {
	Registries []registryEntry `toml:"registries"`
	Compat     int             `toml:"compat"`
}

// LoadConfig reads odin.toml (preferred) or legacy odin.registries.toml from bundlePath.
func LoadConfig(bundlePath string) (*Config, error) {
	if bundlePath == "" {
		bundlePath = "."
	}
	cfg := &Config{Registries: map[string]string{}}

	odinToml := filepath.Join(bundlePath, "odin.toml")
	if st, err := os.Stat(odinToml); err == nil && !st.IsDir() {
		if err := decodeTomlRegistries(odinToml, cfg); err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", odinToml, err)
		}
		return cfg, nil
	}

	return cfg, nil
}

func decodeTomlRegistries(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var root tomlRoot
	if err := toml.NewDecoder(f).Decode(&root); err != nil {
		return err
	}
	for _, r := range root.Registries {
		if r.ModulePrefix == "" || r.Registry == "" {
			continue
		}
		cfg.Registries[r.ModulePrefix] = r.Registry
	}
	cfg.Compat = root.Compat
	return nil
}
