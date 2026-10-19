package ports

import (
	"context"

	"github.com/depsi4/cargos/inventory_service/internal/core/domain"
)

type ProductRepository interface {
	Create(context.Context, string, *domain.Product) error
	Get(context.Context, string, string) (*domain.Product, error)
	List(context.Context, string, int, int) ([]domain.Product, error)
	Update(context.Context, string, *domain.Product) (*domain.Product, error)
	Archive(context.Context, string, string) error
	Adjust(context.Context, string, string, domain.Movement, bool) (*domain.Movement, error)
}
type MovementRepository interface {
	List(context.Context, string, int, int) ([]domain.Movement, error)
}
