package postgres

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// postgres://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable

type DB struct {
	Pool *pgxpool.Pool
}

// Conn - это одиночное соединение с бд
// Pool - набор (пул) готовых соединений, которыми управляет специальный менеджер

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	var (
		err error
		_   = godotenv.Load()
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)

	poolCfg.MinConns = 2
	poolCfg.MaxConns = 10
	poolCfg.HealthCheckPeriod = 30 * time.Second
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping() used for verify that our conn is working
	if err := pool.Ping(pingCtx); err != nil {
		// Close new pool
		pool.Close()
		return nil, err
	}

	log.Println("Database connected")

	return &DB{Pool: pool}, err
}

func (d *DB) Close() {
	d.Pool.Close()
}