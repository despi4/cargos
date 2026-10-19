package postgres

import (
	"context"

	"github.com/depsi4/cargos/auth_service/migrations"
	database "github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct{ Pool *pgxpool.Pool }

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	pool, err := database.Open(ctx, dsn, "auth", migrations.Files)
	if err != nil {
		return nil, err
	}
	return &DB{Pool: pool}, nil
}
func (d *DB) Close() { d.Pool.Close() }
