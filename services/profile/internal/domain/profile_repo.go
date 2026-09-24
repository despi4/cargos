package domain

import (
	"context"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	CreateProfile(ctx context.Context, profile *Profile) (Profile, error)
	GetProfileByID(ctx context.Context, ID uuid.UUID) (*Profile, error)
	GetProfileByUserID(ctx context.Context, UserID uuid.UUID) (*Profile, error)
	UpdateProfile(ctx context.Context, ID uuid.UUID, name string) error
	DeleteProfile(ctx context.Context, ID uuid.UUID) error
}