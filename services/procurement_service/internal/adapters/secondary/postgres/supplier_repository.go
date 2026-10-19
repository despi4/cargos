package postgres

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SupplierRepository struct{ pool *pgxpool.Pool }

func NewSupplierRepository(pool *pgxpool.Pool) *SupplierRepository {
	return &SupplierRepository{pool: pool}
}
func scanSupplier(row postgres.Scanner) (*domain.Supplier, error) {
	var v domain.Supplier
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.Name, &v.Email, &v.LeadDays, &v.Active)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *SupplierRepository) Create(ctx context.Context, owner string, v *domain.Supplier) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO suppliers(id,created_at,updated_at,name,email,lead_days,active,owner_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, v.ID, v.CreatedAt, v.UpdatedAt, v.Name, v.Email, v.LeadDays, v.Active, owner)
	return postgres.Error(err)
}
func (r *SupplierRepository) Get(ctx context.Context, owner, id string) (*domain.Supplier, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanSupplier(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,name,email,lead_days,active FROM suppliers WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *SupplierRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Supplier, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,name,email,lead_days,active FROM suppliers WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Supplier{}
	for rows.Next() {
		v, err := scanSupplier(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}
func (r *SupplierRepository) Update(ctx context.Context, owner string, v *domain.Supplier) (*domain.Supplier, error) {
	return scanSupplier(r.pool.QueryRow(ctx, `UPDATE suppliers SET updated_at=$3,name=$4,email=$5,lead_days=$6 WHERE owner_id=$1 AND id=$2 RETURNING id::text,created_at,updated_at,name,email,lead_days,active`, owner, v.ID, v.UpdatedAt, v.Name, v.Email, v.LeadDays))
}
func (r *SupplierRepository) Archive(ctx context.Context, owner, id string) error {
	if !entity.ValidID(id) {
		return fault.ErrNotFound
	}
	result, err := r.pool.Exec(ctx, `UPDATE suppliers SET active=false,updated_at=now() WHERE owner_id=$1 AND id=$2`, owner, id)
	if err != nil {
		return postgres.Error(err)
	}
	if result.RowsAffected() == 0 {
		return fault.ErrNotFound
	}
	return nil
}
