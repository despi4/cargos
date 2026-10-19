package ports

import (
	"context"

	"github.com/depsi4/cargos/demand_service/internal/core/domain"
)

type Repository interface {
	Create(context.Context, string, *domain.Demand) error
	Get(context.Context, string, string) (*domain.Demand, error)
	List(context.Context, string, int, int) ([]domain.Demand, error)
}
type Inventory interface {
	CheckActive(context.Context, string, string) error
}
