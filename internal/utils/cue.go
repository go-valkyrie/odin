// SPDX-License-Identifier: MIT

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"cuelang.org/go/pkg/strings"
)

func FormatRegistryConfig(registries map[string]string) string {
	r := make([]string, 0, len(registries))
	for path, registry := range registries {
		r = append(r, path+"="+registry)
	}

	slices.Sort(r)

	return strings.Join(r, ",")
}

func CreateCueEnvironment(cacheDir string, registries map[string]string) []string {
	registryConfig := FormatRegistryConfig(registries)
	env := make([]string, 0, 4)

	if v, ok := os.LookupEnv("HOME"); ok {
		env = append(env, "HOME="+v)
	}

	if v, ok := os.LookupEnv("USERPROFILE"); ok {
		env = append(env, "USERPROFILE="+v)
	}

	if cacheDir != "" {
		registrySum := sha256.Sum256([]byte(registryConfig))
		cachePrefix := hex.EncodeToString(registrySum[:])
		env = append(env, fmt.Sprintf("CUE_CACHE_DIR=%s",
			filepath.Join(cacheDir, cachePrefix)))

		env = append(env, fmt.Sprintf("CUE_REGISTRY=%s", registryConfig))
	}

	return env
}
