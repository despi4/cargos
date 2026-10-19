// Package procurementservice is the composition root of the procurement service.
package procurementservice

import (
	"context"
	"net/http"

	client "github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/httpx"
	database "github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/internal/platform/server"
	httpadapter "github.com/depsi4/cargos/procurement_service/internal/adapters/primary/http"
	clients "github.com/depsi4/cargos/procurement_service/internal/adapters/secondary/httpclient"
	"github.com/depsi4/cargos/procurement_service/internal/adapters/secondary/postgres"
	"github.com/depsi4/cargos/procurement_service/internal/core/services"
	"github.com/depsi4/cargos/procurement_service/migrations"
)

func New(ctx context.Context, cfg config.Config) (server.Application, error) {
	pool, err := database.Open(ctx, cfg.DatabaseURL, "procurement", migrations.Files)
	if err != nil {
		return server.Application{}, err
	}
	mux := http.NewServeMux()
	suppliers := postgres.NewSupplierRepository(pool)
	orders := services.NewOrders(postgres.NewOrderRepository(pool), suppliers, &clients.Inventory{Client: client.New(cfg.InternalKey, cfg.URLs)})
	httpadapter.Mount(mux, services.NewSuppliers(suppliers), orders)
	server.Health(mux, "procurement", pool)
	return server.Application{Handler: httpx.Internal(cfg.InternalKey, mux), Close: pool.Close, Tick: orders.RetryReceipts}, nil
}
