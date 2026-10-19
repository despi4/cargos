package domain

import (
	"time"

	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}
type Claims struct {
	UserID    uuid.UUID `json:"sub"`
	SessionID uuid.UUID `json:"sid"`
	ExpiresAt int64     `json:"exp"`
}
type Session struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	RefreshHash string
	ClientIP    string
	UserAgent   string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}
