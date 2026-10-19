package httpclient

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/prediction_service/internal/core/domain"
)

type Sources struct{ Client *client.Client }

func (s *Sources) Products(ctx context.Context, owner string) ([]domain.ProductSnapshot, error) {
	return client.All[domain.ProductSnapshot](ctx, s.Client, owner, "inventory", "/api/v1/products")
}
func (s *Sources) Demands(ctx context.Context, owner string) ([]domain.DemandSnapshot, error) {
	return client.All[domain.DemandSnapshot](ctx, s.Client, owner, "demand", "/api/v1/demand")
}
func (s *Sources) Suppliers(ctx context.Context, owner string) ([]domain.SupplierSnapshot, error) {
	return client.All[domain.SupplierSnapshot](ctx, s.Client, owner, "procurement", "/api/v1/suppliers")
}
func (s *Sources) Orders(ctx context.Context, owner string) ([]domain.OrderSnapshot, error) {
	return client.All[domain.OrderSnapshot](ctx, s.Client, owner, "procurement", "/api/v1/orders")
}
