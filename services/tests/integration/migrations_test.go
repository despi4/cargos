package integration

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	database "github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/inventory_service/migrations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLegacyInventoryMigration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	suffix := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return 'g' + r - '0'
		}
		return r
	}, strings.ReplaceAll(uuid.NewString(), "-", ""))
	schema := "legacy_inventory_" + suffix
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	if _, err = admin.Exec(ctx, "CREATE TABLE "+quoted+".records(id uuid PRIMARY KEY,owner uuid NOT NULL,kind text NOT NULL,payload jsonb NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	owner, id := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	payload, _ := json.Marshal(map[string]any{"id": id, "created_at": now, "updated_at": now, "sku": "LEGACY-1", "name": "Preserved product", "unit": "pcs", "stock": 3, "min_stock": 1, "target_stock": 10, "active": true})
	if _, err = admin.Exec(ctx, "INSERT INTO "+quoted+".records VALUES($1,$2,'product',$3)", id, owner, payload); err != nil {
		t.Fatal(err)
	}
	pool, err := database.Open(ctx, dsn, schema, migrations.Files)
	if err != nil {
		t.Fatal(err)
	}
	var stock int64
	var sku string
	if err = pool.QueryRow(ctx, `SELECT sku,stock FROM products WHERE id=$1 AND owner_id=$2`, id, owner).Scan(&sku, &stock); err != nil {
		t.Fatal(err)
	}
	if sku != "LEGACY-1" || stock != 3 {
		t.Fatalf("legacy product changed: %s stock=%d", sku, stock)
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM records`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("source records were not preserved: count=%d error=%v", count, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE products SET stock=4 WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	pool.Close()
	// Reopening must not reimport the original balance over newer transactions.
	pool, err = database.Open(ctx, dsn, schema, migrations.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = pool.QueryRow(ctx, `SELECT stock FROM products WHERE id=$1`, id).Scan(&stock); err != nil || stock != 4 {
		t.Fatalf("migration replayed: stock=%d error=%v", stock, err)
	}
}
