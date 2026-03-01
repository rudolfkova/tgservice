// Package ...
package tgerror

import "errors"

var (
	// ErrSessionNotFound ...
	ErrSessionNotFound = errors.New("session not found")
	// ErrNotAuthorized ...
	ErrNotAuthorized = errors.New("session is not authorized yet")
	// ErrAlreadyExists ...
	ErrAlreadyExists = errors.New("session already exists")
)

// Wrap ...
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return err
}
