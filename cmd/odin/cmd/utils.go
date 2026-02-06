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
