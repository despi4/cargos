// Package entity contains only common entity metadata, without infrastructure dependencies.
package entity

import (
	"time"

	"github.com/google/uuid"
)

type Base struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func New() Base {
	now := time.Now().UTC()
	return Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}
}
func ValidID(id string) bool { _, err := uuid.Parse(id); return err == nil }
