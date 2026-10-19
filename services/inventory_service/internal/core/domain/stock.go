package domain

import (
	"time"

	"github.com/depsi4/cargos/pkg/fault"
)

const MaxQuantity int64 = 1_000_000_000

func (p *Product) Adjust(quantity int64, receipt bool) error {
	if !p.Active && !receipt {
		return fault.ErrConflict
	}
	if p.Stock+quantity < 0 || p.Stock+quantity > MaxQuantity {
		return fault.ErrConflict
	}
	p.Stock += quantity
	p.UpdatedAt = time.Now().UTC()
	return nil
}
