// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func ensureCacheDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// findBundleRoot walks up from startDir looking for a cue.mod/ directory.
// Returns the absolute path to the bundle root, or an error if none is found.
func findBundleRoot(startDir string) (string, error) {
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	current := absPath
	for {
		// Check if cue.mod exists in current directory
		cueModPath := filepath.Join(current, "cue.mod")
		if info, err := os.Stat(cueModPath); err == nil && info.IsDir() {
			return current, nil
		}

		// Move up one level
		parent := filepath.Dir(current)
		// If we've reached the root, stop
		if parent == current {
			return "", fmt.Errorf("not inside a CUE module (no cue.mod directory found)")
		}
		current = parent
	}
}
