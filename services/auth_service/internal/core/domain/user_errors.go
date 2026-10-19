package domain

import (
	"errors"

	"github.com/depsi4/cargos/pkg/fault"
)

var (
	ErrUserNotFound       = fault.ErrNotFound
	ErrInvalidCredentials = fault.ErrUnauthorized
	ErrUserAlreadyExists  = fault.ErrConflict
	ErrUnauthorized       = fault.ErrUnauthorized
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidArgument    = errors.New("invalid argument")
)
