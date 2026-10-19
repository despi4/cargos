package postgres

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/inventory_service/internal/core/domain"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovementRepository struct{ pool *pgxpool.Pool }

func NewMovementRepository(pool *pgxpool.Pool) *MovementRepository {
	return &MovementRepository{pool: pool}
}
func scanMovement(row postgres.Scanner) (*domain.Movement, error) {
	var v domain.Movement
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.ProductID, &v.Quantity, &v.Reason, &v.Reference, &v.Balance)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *MovementRepository) Get(ctx context.Context, owner, id string) (*domain.Movement, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanMovement(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,product_id::text,quantity,reason,reference,balance FROM movements WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *MovementRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Movement, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,product_id::text,quantity,reason,reference,balance FROM movements WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Movement{}
	for rows.Next() {
		v, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}
