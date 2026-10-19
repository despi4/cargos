package httpclient

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/pkg/fault"
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
