package services

import (
	"context"
	"errors"

	"github.com/depsi4/cargos/notification_service/internal/core/domain"
	"github.com/depsi4/cargos/notification_service/internal/core/ports"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
)

type Notifications struct{ repository ports.Repository }

func New(repo ports.Repository) *Notifications { return &Notifications{repository: repo} }

// Publish returns nil for an event already delivered. The unique event ID makes retries safe.
func (s *Notifications) Publish(ctx context.Context, owner, eventID, title, message string) (*domain.Notification, error) {
	if !entity.ValidID(eventID) || len(title) < 1 || len(title) > 200 || len(message) < 1 || len(message) > 2000 {
		return nil, fault.Invalid("valid event_id, title and message required")
	}
	v := &domain.Notification{Base: entity.New(), EventID: eventID, Title: title, Message: message}
	err := s.repository.Create(ctx, owner, v)
	if errors.Is(err, fault.ErrConflict) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Notifications) List(ctx context.Context, owner string, limit, offset int) ([]domain.Notification, error) {
	return s.repository.List(ctx, owner, limit, offset)
}

func (s *Notifications) Get(ctx context.Context, owner, id string) (*domain.Notification, error) {
	return s.repository.Get(ctx, owner, id)
}

func (s *Notifications) SetRead(ctx context.Context, owner, id string, read bool) (*domain.Notification, error) {
	return s.repository.SetRead(ctx, owner, id, read)
}
