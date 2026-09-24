package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ConnDB struct {
	db *sql.DB
}

func NewPostgresDB(ctx context.Context, dsn string) (*ConnDB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to load driver: %w", err)
	}

	// Настройка Connection Pool
	db.SetMaxOpenConns(25)                 // Максимум открытых соединений
	db.SetMaxIdleConns(25)                 // Максимум простаивающих соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Время жизни соединения
	db.SetConnMaxIdleTime(5 * time.Minute) // Время жизни простаивающего соединения

	// Проверяем, что база реально доступна
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &ConnDB{db: db}, nil
}

func (c *ConnDB) Close() error {
	return c.db.Close()
}

func (c *ConnDB) GetDB() *sql.DB {
	return c.db
}
