package ports

import (
	"context"

	"github.com/depsi4/cargos/prediction_service/internal/core/domain"
)

type Repository interface {
	Create(context.Context, string, *domain.Prediction) error
	Get(context.Context, string, string) (*domain.Prediction, error)
	List(context.Context, string, int, int) ([]domain.Prediction, error)
	Pending(context.Context) ([]domain.PendingDelivery, error)
	MarkSent(context.Context, string, string) error
}
type Source interface {
	Products(context.Context, string) ([]domain.ProductSnapshot, error)
	Demands(context.Context, string) ([]domain.DemandSnapshot, error)
	Suppliers(context.Context, string) ([]domain.SupplierSnapshot, error)
	Orders(context.Context, string) ([]domain.OrderSnapshot, error)
}
type Notifier interface {
	Publish(context.Context, string, string, string, string) error
}
