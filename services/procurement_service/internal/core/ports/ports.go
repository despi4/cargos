package ports

import (
	"context"

	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
)

type SupplierRepository interface {
	Create(context.Context, string, *domain.Supplier) error
	Get(context.Context, string, string) (*domain.Supplier, error)
	List(context.Context, string, int, int) ([]domain.Supplier, error)
	Update(context.Context, string, *domain.Supplier) (*domain.Supplier, error)
	Archive(context.Context, string, string) error
}
type OrderRepository interface {
	Create(context.Context, string, *domain.Order) error
	Get(context.Context, string, string) (*domain.Order, error)
	List(context.Context, string, int, int) ([]domain.Order, error)
	Update(context.Context, string, string, func(*domain.Order) error) (*domain.Order, error)
	Pending(context.Context) ([]domain.PendingReceipt, error)
}
type Inventory interface {
	CheckActive(context.Context, string, string) error
	Receive(context.Context, string, *domain.Order) error
}
