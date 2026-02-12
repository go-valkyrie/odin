// SPDX-License-Identifier: MIT

package config

type ErrConfigValidation struct {
	wrapped error
}

func (e *ErrConfigValidation) Error() string {
	return "config validation failed"
}

func (e *ErrConfigValidation) Unwrap() error {
	return e.wrapped
}
