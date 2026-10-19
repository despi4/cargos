// Package fault defines application errors independently of HTTP and PostgreSQL.
package fault

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists or conflicts with current state")
	ErrUnauthorized = errors.New("invalid or expired credentials")
)

type InvalidInput struct{ Message string }

func (e *InvalidInput) Error() string { return e.Message }
func Invalid(message string) error    { return &InvalidInput{Message: message} }

type Unavailable struct{ Message string }

func (e *Unavailable) Error() string { return e.Message }
