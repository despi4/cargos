package postgres

import (
	"context"
	"errors"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/inventory_service/internal/core/domain"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
)

// Adjust commits the locked balance and its unique ledger entry in one transaction.
func (r *ProductRepository) Adjust(ctx context.Context, owner, id string, m domain.Movement, receipt bool) (*domain.Movement, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	p, err := scanProduct(tx.QueryRow(ctx, `SELECT id::text,created_at,updated_at,sku,name,unit,stock,min_stock,target_stock,active FROM products WHERE owner_id=$1 AND id=$2 FOR UPDATE`, owner, id))
	if err != nil {
		return nil, err
	}
	prior, err := scanMovement(tx.QueryRow(ctx, `SELECT id::text,created_at,updated_at,product_id::text,quantity,reason,reference,balance FROM movements WHERE owner_id=$1 AND reference=$2`, owner, m.Reference))
	if err == nil {
		if prior.ProductID != id || prior.Quantity != m.Quantity || prior.Reason != m.Reason {
			return nil, fault.ErrConflict
		}
		return prior, nil
	}
	if !errors.Is(err, fault.ErrNotFound) {
		return nil, err
	}
	if err = p.Adjust(m.Quantity, receipt); err != nil {
		return nil, err
	}
	m.Balance = p.Stock
	_, err = tx.Exec(ctx, `INSERT INTO movements(id,owner_id,created_at,updated_at,product_id,quantity,reason,reference,balance) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, m.ID, owner, m.CreatedAt, m.UpdatedAt, id, m.Quantity, m.Reason, m.Reference, m.Balance)
	if err != nil {
		return nil, postgres.Error(err)
	}
	_, err = tx.Exec(ctx, `UPDATE products SET stock=$3,updated_at=$4 WHERE owner_id=$1 AND id=$2`, owner, id, p.Stock, p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &m, nil
}
