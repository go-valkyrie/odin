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

package initialize

import (
	"bytes"
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/mod/modfile"
	"cuelang.org/go/pkg/strings"
	"fmt"
	giturls "github.com/chainguard-dev/git-urls"
	"go-valkyrie.com/odin/internal/git"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func (o *Options) Run(ctx context.Context) error {
	return run(ctx, o)
}

func bundleNameFromGit(bundleName string) (string, error) {
	repo, err := git.OpenPath(bundleName)
	if err != nil {
		return "", err
	}

	var originURL *url.URL
	if origin, ok := git.OriginPath(repo); !ok {
		originURL = nil
	} else if origin, err := giturls.Parse(origin); err != nil {
		return "", err
	} else {
		originURL = origin
	}

	if originURL == nil {
		return "", fmt.Errorf("unable to determine bundle name from git repository")
	}

	path := originURL.EscapedPath()
	parts := strings.Split(path, "/")
	lastPart := parts[len(parts)-1]

	return strings.TrimSuffix(lastPart, ".git"), nil
}

func moduleNameFromGit(bundlePath string) (string, error) {
	repo, err := git.OpenPath(bundlePath)
	if err != nil {
		return "", err
	}

	var originURL *url.URL
	if origin, ok := git.OriginPath(repo); !ok {
		originURL = nil
	} else if origin, err := giturls.Parse(origin); err != nil {
		return "", err
	} else {
		originURL = origin
	}

	if originURL == nil {
		return "", fmt.Errorf("unable to determine module name from git repository")
	}

	modulePath := fmt.Sprintf("%s/%s", originURL.Hostname(), originURL.EscapedPath())

	repoRoot, err := git.RepoRoot(repo)
	if err != nil {
		return "", err
	}

	bundlePath, err = filepath.Abs(bundlePath)
	if err != nil {
		return "", err
	}

	relpath, err := filepath.Rel(repoRoot, bundlePath)
	if err != nil {
		return "", err
	}

	if relpath == "." {
		return modulePath, nil
	} else {
		return url.JoinPath(modulePath, relpath)
	}
}

func run(ctx context.Context, o *Options) error {
	logger := o.Logger

	bundlePath := o.BundlePath
	if path, err := filepath.Abs(bundlePath); err != nil {
		return err
	} else {
		bundlePath = path
	}

	if fi, err := os.Stat(bundlePath); err != nil && os.IsNotExist(err) {
		logger.InfoContext(ctx, "directory does not exist, creating it", "path", bundlePath)
		err := os.MkdirAll(bundlePath, 0755)
		if err != nil {
			return fmt.Errorf("path for new bundhle does not exist and unable to create: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("unable to determine if path for new bundle exists: %w", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("path for new bundle exists and is a file, exiting")
	} else if entries, err := os.ReadDir(bundlePath); err != nil {
		return fmt.Errorf("unable to read directory for new bundle: %w", err)
	} else if len(entries) > 0 {
		return fmt.Errorf("path for new bundle exists and is not empty, exiting")
	}

	modulePath := o.ModulePath
	if modulePath == "" {
		if nameFromGit, err := moduleNameFromGit(o.BundlePath); err == nil {
			modulePath = nameFromGit
		} else {
			return fmt.Errorf("module path not specified and unable to determine from git repository: %w", err)
		}
	}

	bundleName := o.BundleName
	if bundleName == "" {
		if nameFromGit, err := bundleNameFromGit(o.BundlePath); err == nil {
			bundleName = nameFromGit
		} else {
			return fmt.Errorf("bundle name not specified and unable to determine from git repository: %w", err)
		}
	}

	packageName := ""
	if components := strings.Split(modulePath, "/"); len(components) < 2 {
		return fmt.Errorf("invalid module name, must have a path")
	} else {
		packageName = components[len(components)-1]
	}

	cueModulePath := filepath.Join(bundlePath, "cue.mod", "module.cue")
	if fi, err := os.Stat(cueModulePath); err == nil && !fi.IsDir() {
		return fmt.Errorf("module.cue already exists in bundle path")
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to determine if module.cue already exists in bundle path: %w", err)
	}

	logger.InfoContext(ctx, "creating bundle", "module", modulePath, "path", bundlePath)

	cueModFile := modfile.File{
		Module: modulePath,
		Language: &modfile.Language{
			Version: cue.LanguageVersion(),
		},
	}

	if err := os.Mkdir(filepath.Join(bundlePath, "cue.mod"), 0755); err != nil {
		return err
	}

	failedInit := true
	defer func() {
		if failedInit {
			entries, err := filepath.Glob(filepath.Join(bundlePath, "*"))
			if err != nil {

			}
			for _, entry := range entries {
				os.RemoveAll(entry)
			}
		}
	}()

	if data, err := cueModFile.Format(); err != nil {
		return err
	} else if err := os.WriteFile(cueModulePath, data, 0644); err != nil {
		return err
	}

	if template := bundleTemplate.Lookup("bundle.cue.tmpl"); template == nil {
		return fmt.Errorf("unable to find template for bundle")
	} else {
		var buffer bytes.Buffer
		buffer.WriteString("package ")
		buffer.WriteString(packageName)
		buffer.WriteString("\n\n")
		if err := template.Execute(&buffer, struct {
			BundleName  string
			PackageName string
		}{
			BundleName:  bundleName,
			PackageName: packageName,
		}); err != nil {
			return fmt.Errorf("unable to execute template for bundle: %w", err)
		}
		err := os.WriteFile(filepath.Join(bundlePath, "bundle.cue"), buffer.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	c := exec.Command(os.Args[0], "cue", "mod", "tidy")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = bundlePath
	if err := c.Run(); err != nil {
		return err
	}

	c = exec.Command(os.Args[0], "cue", "fmt")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = bundlePath
	if err := c.Run(); err != nil {
		return err
	}

	failedInit = false
	return nil
}
