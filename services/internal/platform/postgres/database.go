// Package postgres provides connection and migration infrastructure only.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/depsi4/cargos/pkg/fault"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, dsn, schema string, migrations fs.FS) (*pgxpool.Pool, error) {
	for _, c := range schema {
		if (c < 'a' || c > 'z') && c != '_' {
			return nil, fmt.Errorf("invalid database schema")
		}
	}
	if schema == "" {
		return nil, fmt.Errorf("database schema required")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			pool.Close()
		}
	}()
	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "cargos-migrate-"+schema); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS `+schema+`;CREATE TABLE IF NOT EXISTS schema_migrations(name TEXT PRIMARY KEY,applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return nil, err
	}
	names, err := fs.Glob(migrations, "*.up.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name=$1)`, name).Scan(&exists); err != nil {
			return nil, err
		}
		if exists {
			continue
		}
		data, err := fs.ReadFile(migrations, name)
		if err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, string(data)); err != nil {
			return nil, fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(name) VALUES($1)`, name); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	success = true
	return pool, nil
}
func Error(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fault.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return fault.ErrConflict
		}
		if pgErr.Code == "22P02" && strings.Contains(pgErr.Message, "uuid") {
			return fault.ErrNotFound
		}
	}
	return err
}

type Scanner interface{ Scan(...any) error }
