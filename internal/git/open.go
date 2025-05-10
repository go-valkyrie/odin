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

package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
)

type Repository = git.Repository

func OpenPath(path string) (*Repository, error) {
	if repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	}); err != nil {
		return nil, &ErrCannotOpenRepo{Path: path, wrapped: err}
	} else if config, err := repo.Config(); err != nil {
		return nil, fmt.Errorf("unable to get config for repo: %w", err)
	} else if config.Core.IsBare {
		return nil, &ErrCannotUseBareRepo{Path: path}
	} else {
		return repo, nil
	}
}

func RepoRoot(repo *Repository) (string, error) {
	if wt, err := repo.Worktree(); err != nil {
		return "", err
	} else {
		return wt.Filesystem.Root(), nil
	}
}

func OriginPath(repo *Repository) (string, bool) {
	if remote, err := repo.Remote("origin"); err != nil {
		return "", false
	} else {
		return remote.Config().URLs[0], true
	}
}
