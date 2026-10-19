package postgres

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct{ pool *pgxpool.Pool }

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository { return &OrderRepository{pool: pool} }
func scanOrder(row postgres.Scanner) (*domain.Order, error) {
	var v domain.Order
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.ProductID, &v.SupplierID, &v.Quantity, &v.UnitPrice, &v.Currency, &v.Status, &v.ExpectedAt)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *OrderRepository) Create(ctx context.Context, owner string, v *domain.Order) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO orders(id,created_at,updated_at,product_id,supplier_id,quantity,unit_price,currency,status,expected_at,owner_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, v.ID, v.CreatedAt, v.UpdatedAt, v.ProductID, v.SupplierID, v.Quantity, v.UnitPrice, v.Currency, v.Status, v.ExpectedAt, owner)
	return postgres.Error(err)
}
func (r *OrderRepository) Get(ctx context.Context, owner, id string) (*domain.Order, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanOrder(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,product_id::text,supplier_id::text,quantity,unit_price,currency,status,expected_at FROM orders WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *OrderRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,product_id::text,supplier_id::text,quantity,unit_price,currency,status,expected_at FROM orders WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Order{}
	for rows.Next() {
		v, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}

// Update serializes state transitions and commits the mutation atomically.
func (r *OrderRepository) Update(ctx context.Context, owner, id string, mutate func(*domain.Order) error) (*domain.Order, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	v, err := scanOrder(tx.QueryRow(ctx, `SELECT id::text,created_at,updated_at,product_id::text,supplier_id::text,quantity,unit_price,currency,status,expected_at FROM orders WHERE owner_id=$1 AND id=$2 FOR UPDATE`, owner, id))
	if err != nil {
		return nil, err
	}
	if err = mutate(v); err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `UPDATE orders SET created_at=$3,updated_at=$4,product_id=$5,supplier_id=$6,quantity=$7,unit_price=$8,currency=$9,status=$10,expected_at=$11 WHERE owner_id=$1 AND id=$2`, owner, id, v.CreatedAt, v.UpdatedAt, v.ProductID, v.SupplierID, v.Quantity, v.UnitPrice, v.Currency, v.Status, v.ExpectedAt)
	if err != nil {
		return nil, postgres.Error(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return v, nil
}
func (r *OrderRepository) Pending(ctx context.Context) ([]domain.PendingReceipt, error) {
	rows, err := r.pool.Query(ctx, `SELECT owner_id::text,id::text FROM orders WHERE status='receiving' ORDER BY created_at LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.PendingReceipt{}
	for rows.Next() {
		var v domain.PendingReceipt
		if err = rows.Scan(&v.OwnerID, &v.OrderID); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}
