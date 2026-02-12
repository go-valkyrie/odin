// SPDX-License-Identifier: MIT

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
