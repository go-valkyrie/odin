// Copyright 2025 Stefan Nuxoll
// SPDX-License-Identifier: MIT

package oci

import (
	"testing"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantReg    string
		wantRepo   string
		wantRef    string
		wantErr    bool
		wantString string
	}{
		{
			name:       "basic tag",
			input:      "ghcr.io/org/app:v1",
			wantReg:    "ghcr.io",
			wantRepo:   "org/app",
			wantRef:    "v1",
			wantString: "ghcr.io/org/app:v1",
		},
		{
			name:       "with oci:// scheme",
			input:      "oci://ghcr.io/org/app:v1",
			wantReg:    "ghcr.io",
			wantRepo:   "org/app",
			wantRef:    "v1",
			wantString: "ghcr.io/org/app:v1",
		},
		{
			name:       "no tag defaults to latest",
			input:      "ghcr.io/org/app",
			wantReg:    "ghcr.io",
			wantRepo:   "org/app",
			wantRef:    "latest",
			wantString: "ghcr.io/org/app:latest",
		},
		{
			name:       "digest reference",
			input:      "ghcr.io/org/app@sha256:abcdef",
			wantReg:    "ghcr.io",
			wantRepo:   "org/app",
			wantRef:    "sha256:abcdef",
			wantString: "ghcr.io/org/app@sha256:abcdef",
		},
		{
			name:       "nested repository",
			input:      "registry.example.com/org/project/bundle:tag",
			wantReg:    "registry.example.com",
			wantRepo:   "org/project/bundle",
			wantRef:    "tag",
			wantString: "registry.example.com/org/project/bundle:tag",
		},
		{
			name:    "invalid - no repository",
			input:   "ghcr.io",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseReference(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if ref.Registry != tt.wantReg {
				t.Errorf("Registry = %v, want %v", ref.Registry, tt.wantReg)
			}
			if ref.Repository != tt.wantRepo {
				t.Errorf("Repository = %v, want %v", ref.Repository, tt.wantRepo)
			}
			if ref.Reference != tt.wantRef {
				t.Errorf("Reference = %v, want %v", ref.Reference, tt.wantRef)
			}
			if ref.String() != tt.wantString {
				t.Errorf("String() = %v, want %v", ref.String(), tt.wantString)
			}
		})
	}
}

func TestReferenceLastComponent(t *testing.T) {
	tests := []struct {
		name string
		ref  *Reference
		want string
	}{
		{
			name: "simple repo",
			ref:  &Reference{Repository: "org/app"},
			want: "app",
		},
		{
			name: "nested repo",
			ref:  &Reference{Repository: "org/project/bundle"},
			want: "bundle",
		},
		{
			name: "single component",
			ref:  &Reference{Repository: "app"},
			want: "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ref.LastComponent(); got != tt.want {
				t.Errorf("LastComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}
