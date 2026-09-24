package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID uuid.UUID
	UserID uuid.UUID
	Name string
	Age *int
	Skills []string
	Positions []string
	Location string
	Experience int
	Education string
	Preferences []string
	CreatedAt time.Time
	UpdatedAt time.Time
}