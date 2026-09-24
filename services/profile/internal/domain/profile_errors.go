package domain

import "errors"

var (
	ErrProfileNotFound        = errors.New("profile not found")
	ErrProfileExists          = errors.New("profile already exists")
)