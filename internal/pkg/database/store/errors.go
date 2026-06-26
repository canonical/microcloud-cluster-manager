package store

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates a requested store entry does not exist.
	ErrNotFound = errors.New("store: not found")
	// ErrAlreadyExists indicates a store entry already exists.
	ErrAlreadyExists = errors.New("store: already exists")
)

// NotFoundErrorf returns an ErrNotFound-wrapped error with context.
func NotFoundErrorf(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrNotFound, fmt.Sprintf(format, a...))
}

// AlreadyExistsErrorf returns an ErrAlreadyExists-wrapped error with context.
func AlreadyExistsErrorf(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrAlreadyExists, fmt.Sprintf(format, a...))
}

// IsNotFound reports whether err wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists reports whether err wraps ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
