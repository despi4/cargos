package migrator

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(fs embed.FS, dir string, dsn string) error {
	d, err := iofs.New(fs, dir)
	if err != nil {
		return fmt.Errorf("create iofs driver error: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("init migrator error: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations errors: %w", err)
	}

	log.Println("Миграции успешно применены!")

	return nil
}
