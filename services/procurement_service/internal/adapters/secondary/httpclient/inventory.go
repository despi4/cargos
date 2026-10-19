package httpclient

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
)

type Inventory struct{ Client *client.Client }

func (i *Inventory) CheckActive(ctx context.Context, owner, id string) error {
	var product struct {
		Active bool `json:"active"`
	}
	if err := i.Client.Do(ctx, owner, "inventory", "GET", "/api/v1/products/"+id, nil, &product); err != nil {
		return err
	}
	if !product.Active {
		return fault.ErrConflict
	}
	return nil
}
func (i *Inventory) Receive(ctx context.Context, owner string, order *domain.Order) error {
	return i.Client.Do(ctx, owner, "inventory", "POST", "/internal/products/"+order.ProductID+"/receipts", map[string]any{"quantity": order.Quantity, "reason": "purchase received", "reference": "purchase:" + order.ID}, nil)
}
