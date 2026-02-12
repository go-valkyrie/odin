// SPDX-License-Identifier: MIT

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatRegistryConfig(t *testing.T) {
	tests := []struct {
		name       string
		registries map[string]string
		want       string
	}{
		{
			name:       "empty map",
			registries: map[string]string{},
			want:       "",
		},
		{
			name: "single entry",
			registries: map[string]string{
				"example.com": "registry.example.com",
			},
			want: "example.com=registry.example.com",
		},
		{
			name: "multiple entries sorted",
			registries: map[string]string{
				"z.com": "registry.z.com",
				"a.com": "registry.a.com",
				"m.com": "registry.m.com",
			},
			want: "a.com=registry.a.com,m.com=registry.m.com,z.com=registry.z.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRegistryConfig(tt.registries)
			if got != tt.want {
				t.Errorf("FormatRegistryConfig() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateCueEnvironment(t *testing.T) {
	// Save and restore original env vars
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	defer func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		} else {
			os.Unsetenv("HOME")
		}
		if origUserProfile != "" {
			os.Setenv("USERPROFILE", origUserProfile)
		} else {
			os.Unsetenv("USERPROFILE")
		}
	}()

	tests := []struct {
		name            string
		cacheDir        string
		registries      map[string]string
		setHome         bool
		setUserProfile  bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:       "no cache dir",
			cacheDir:   "",
			registries: map[string]string{"example.com": "registry.example.com"},
			setHome:    true,
			wantContains: []string{
				"HOME=",
			},
			wantNotContains: []string{
				"CUE_CACHE_DIR=",
				"CUE_REGISTRY=",
			},
		},
		{
			name:       "with cache dir",
			cacheDir:   "/tmp/cache",
			registries: map[string]string{"example.com": "registry.example.com"},
			setHome:    true,
			wantContains: []string{
				"HOME=",
				"CUE_CACHE_DIR=",
				"CUE_REGISTRY=example.com=registry.example.com",
			},
		},
		{
			name:       "cache dir uses sha256 prefix",
			cacheDir:   "/tmp/cache",
			registries: map[string]string{"a.com": "reg.a.com"},
			setHome:    true,
			wantContains: []string{
				"CUE_CACHE_DIR=/tmp/cache/",
			},
		},
		{
			name:           "with userprofile",
			cacheDir:       "",
			registries:     map[string]string{},
			setUserProfile: true,
			wantContains: []string{
				"USERPROFILE=",
			},
		},
		{
			name:       "no HOME or USERPROFILE set",
			cacheDir:   "",
			registries: map[string]string{},
			wantNotContains: []string{
				"HOME=",
				"USERPROFILE=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup env vars
			if tt.setHome {
				os.Setenv("HOME", "/home/testuser")
			} else {
				os.Unsetenv("HOME")
			}
			if tt.setUserProfile {
				os.Setenv("USERPROFILE", "C:\\Users\\testuser")
			} else {
				os.Unsetenv("USERPROFILE")
			}

			got := CreateCueEnvironment(tt.cacheDir, tt.registries)
			gotStr := strings.Join(got, "\n")

			for _, want := range tt.wantContains {
				found := false
				for _, env := range got {
					if strings.Contains(env, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("CreateCueEnvironment() missing %q; got %v", want, got)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(gotStr, notWant) {
					t.Errorf("CreateCueEnvironment() should not contain %q; got %v", notWant, got)
				}
			}

			// Verify cache dir includes sha256 hash prefix
			if tt.cacheDir != "" {
				registryConfig := FormatRegistryConfig(tt.registries)
				registrySum := sha256.Sum256([]byte(registryConfig))
				expectedPrefix := hex.EncodeToString(registrySum[:])
				expectedCacheDir := filepath.Join(tt.cacheDir, expectedPrefix)

				found := false
				for _, env := range got {
					if strings.HasPrefix(env, "CUE_CACHE_DIR=") && strings.Contains(env, expectedPrefix) {
						found = true
						if !strings.HasSuffix(env, expectedCacheDir) {
							t.Errorf("CUE_CACHE_DIR = %q, want to end with %q", env, expectedCacheDir)
						}
						break
					}
				}
				if !found {
					t.Errorf("CreateCueEnvironment() missing CUE_CACHE_DIR with sha256 prefix %q", expectedPrefix)
				}
			}
		})
	}
}
