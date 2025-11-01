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

package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type config struct {
	Registries map[string]string
}

type registryEntry struct {
	ModulePrefix string `toml:"module-prefix"`
	Registry     string `toml:"registry"`
}

type tomlRoot struct {
	Registries []registryEntry `toml:"registries"`
}

// loadConfig reads odin.toml (preferred) or legacy odin.registries.toml from bundlePath.
func loadConfig(bundlePath string) (*config, error) {
	if bundlePath == "" {
		bundlePath = "."
	}
	cfg := &config{Registries: map[string]string{}}

	odinToml := filepath.Join(bundlePath, "odin.toml")
	if st, err := os.Stat(odinToml); err == nil && !st.IsDir() {
		if err := decodeTomlRegistries(odinToml, cfg); err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", odinToml, err)
		}
		return cfg, nil
	}

	return cfg, nil
}

func decodeTomlRegistries(path string, cfg *config) error {
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
