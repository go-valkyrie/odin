// SPDX-License-Identifier: MIT

package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Registries map[string]string
}

type registryEntry struct {
	ModulePrefix string `toml:"module-prefix"`
	Registry     string `toml:"registry"`
}

type tomlRoot struct {
	Registries []registryEntry `toml:"registries"`
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
	return nil
}
