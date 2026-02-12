// SPDX-License-Identifier: MIT

package git

import "fmt"

type ErrCannotOpenRepo struct {
	Path    string
	wrapped error
}

func (e *ErrCannotOpenRepo) Unwrap() error {
	return e.wrapped
}

func (e *ErrCannotOpenRepo) Error() string {
	return fmt.Sprintf("unable to open repository at '%s': %s", e.Path, e.wrapped.Error())
}

type ErrCannotUseBareRepo struct {
	Path string
}

func (e *ErrCannotUseBareRepo) Error() string {
	return fmt.Sprintf("repository at '%s' is a bare repository and cannot be used", e.Path)
}
