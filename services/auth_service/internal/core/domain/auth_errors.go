package domain

import "errors"

var (
	ErrEmailTooShort      = errors.New("email is too short (min 6 characters)")
	ErrEmailTooLong       = errors.New("email is too long (max 254 characters)")
	ErrEmailInvalidFormat = errors.New("invalid email format")

	ErrPasswordTooShort   = errors.New("password is too short (min 8 characters)")
	ErrPasswordTooLong    = errors.New("password is too long (max 32 characters)")
	ErrPasswordNoUpper    = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower    = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit    = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial  = errors.New("password must contain at least one special character")

	ErrUsernameTooShort   = errors.New("username is too short (min 3 characters)")
	ErrUsernameTooLong    = errors.New("username is too long (max 25 characters)")

	ErrInvalidClientIP     = errors.New("client IP is required")
	ErrInvalidUserAgent    = errors.New("user agent is required")
	ErrInvalidRefreshToken = errors.New("refresh token is required")

	ErrSessionBlocked  = errors.New("session is blocked")
	ErrSessionExpired  = errors.New("session is expired")
	ErrSessionNotFound = errors.New("session not found")
)