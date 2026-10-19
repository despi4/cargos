package postgres

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/inventory_service/internal/core/domain"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct{ pool *pgxpool.Pool }

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}
func scanProduct(row postgres.Scanner) (*domain.Product, error) {
	var v domain.Product
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.SKU, &v.Name, &v.Unit, &v.Stock, &v.MinStock, &v.TargetStock, &v.Active)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *ProductRepository) Create(ctx context.Context, owner string, v *domain.Product) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO products(id,created_at,updated_at,sku,name,unit,stock,min_stock,target_stock,active,owner_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, v.ID, v.CreatedAt, v.UpdatedAt, v.SKU, v.Name, v.Unit, v.Stock, v.MinStock, v.TargetStock, v.Active, owner)
	return postgres.Error(err)
}
func (r *ProductRepository) Get(ctx context.Context, owner, id string) (*domain.Product, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanProduct(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,sku,name,unit,stock,min_stock,target_stock,active FROM products WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *ProductRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,sku,name,unit,stock,min_stock,target_stock,active FROM products WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Product{}
	for rows.Next() {
		v, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}
func (r *ProductRepository) Update(ctx context.Context, owner string, v *domain.Product) (*domain.Product, error) {
	return scanProduct(r.pool.QueryRow(ctx, `UPDATE products SET updated_at=$3,name=$4,unit=$5,min_stock=$6,target_stock=$7 WHERE owner_id=$1 AND id=$2 RETURNING id::text,created_at,updated_at,sku,name,unit,stock,min_stock,target_stock,active`, owner, v.ID, v.UpdatedAt, v.Name, v.Unit, v.MinStock, v.TargetStock))
}
func (r *ProductRepository) Archive(ctx context.Context, owner, id string) error {
	if !entity.ValidID(id) {
		return fault.ErrNotFound
	}
	result, err := r.pool.Exec(ctx, `UPDATE products SET active=false,updated_at=now() WHERE owner_id=$1 AND id=$2`, owner, id)
	if err != nil {
		return postgres.Error(err)
	}
	if result.RowsAffected() == 0 {
		return fault.ErrNotFound
	}
	return nil
}
