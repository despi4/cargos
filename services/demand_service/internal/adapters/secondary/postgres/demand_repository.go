package postgres

import (
	"context"

	"github.com/depsi4/cargos/demand_service/internal/core/domain"
	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DemandRepository struct{ pool *pgxpool.Pool }

func NewDemandRepository(pool *pgxpool.Pool) *DemandRepository { return &DemandRepository{pool: pool} }
func scanDemand(row postgres.Scanner) (*domain.Demand, error) {
	var v domain.Demand
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.ProductID, &v.Quantity, &v.OccurredAt, &v.Note)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *DemandRepository) Create(ctx context.Context, owner string, v *domain.Demand) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO demand_records(id,created_at,updated_at,product_id,quantity,occurred_at,note,owner_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, v.ID, v.CreatedAt, v.UpdatedAt, v.ProductID, v.Quantity, v.OccurredAt, v.Note, owner)
	return postgres.Error(err)
}
func (r *DemandRepository) Get(ctx context.Context, owner, id string) (*domain.Demand, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanDemand(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,product_id::text,quantity,occurred_at,note FROM demand_records WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *DemandRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Demand, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,product_id::text,quantity,occurred_at,note FROM demand_records WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Demand{}
	for rows.Next() {
		v, err := scanDemand(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}
