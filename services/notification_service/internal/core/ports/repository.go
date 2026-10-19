package ports

import (
	"context"

	"github.com/depsi4/cargos/notification_service/internal/core/domain"
)

type Repository interface {
	Create(context.Context, string, *domain.Notification) error
	Get(context.Context, string, string) (*domain.Notification, error)
	List(context.Context, string, int, int) ([]domain.Notification, error)
	SetRead(context.Context, string, string, bool) (*domain.Notification, error)
}
