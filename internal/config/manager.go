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

package config

import (
	"cuelang.org/go/cue"
	cuefmt "cuelang.org/go/cue/format"
	"fmt"
	"go-valkyrie.com/cueconfig"
	"log/slog"
	"sync"
)

// Manager is the interface for configuration management
type Manager interface {
	Evaluated() ([]byte, error)
	Load() error
	ModuleRegistries() (map[string]string, error)
	Raw() *cue.Value
}

// manager is a thin wrapper around cueconfig.Config
type manager struct {
	config     *cueconfig.Config
	configMu   sync.Mutex
	configPath string
	logger     *slog.Logger
}

// NewManager creates a new configuration manager
func NewManager(logger *slog.Logger, configPath string) (Manager, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	} else {
		logger = logger.With("component", "config")
	}

	// Create config with schema and source
	configSource, err := loadSource(configPath)
	if err != nil {
		return nil, err
	}

	config, err := cueconfig.New(configSchema, cueconfig.WithSources(configSource))
	if err != nil {
		return nil, err
	}

	return &manager{
		config:     config,
		configPath: configPath,
		logger:     logger,
	}, nil
}

// Load reloads the configuration
func (m *manager) Load() error {
	m.configMu.Lock()
	defer m.configMu.Unlock()

	return m.config.Reload()
}

// Evaluated returns the evaluated configuration as formatted bytes
func (m *manager) Evaluated() ([]byte, error) {
	if err := m.Load(); err != nil {
		return nil, err
	}

	// Get the raw CUE value and format it
	value := m.config.Raw().Eval()
	syntax := value.Syntax(cue.Final())
	return cuefmt.Node(syntax, cuefmt.Simplify())
}

// ModuleRegistries returns the module registries from the configuration
func (m *manager) ModuleRegistries() (map[string]string, error) {
	registries := make(map[string]string)
	if err := m.config.ValueAt("cue.registries").Decode(&registries); err != nil {
		return nil, err
	}
	return registries, nil
}

// Raw returns the raw CUE value
func (m *manager) Raw() *cue.Value {
	return m.config.Raw()
}
