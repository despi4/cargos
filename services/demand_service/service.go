// Package demandservice is the composition root of the demand service.
package demandservice

import (
	"context"
	"net/http"

	httpadapter "github.com/depsi4/cargos/demand_service/internal/adapters/primary/http"
	clients "github.com/depsi4/cargos/demand_service/internal/adapters/secondary/httpclient"
	"github.com/depsi4/cargos/demand_service/internal/adapters/secondary/postgres"
	"github.com/depsi4/cargos/demand_service/internal/core/services"
	"github.com/depsi4/cargos/demand_service/migrations"
	client "github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/httpx"
	database "github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/internal/platform/server"
)

func New(ctx context.Context, cfg config.Config) (server.Application, error) {
	pool, err := database.Open(ctx, cfg.DatabaseURL, "demand", migrations.Files)
	if err != nil {
		return server.Application{}, err
	}
	mux := http.NewServeMux()
	s := services.New(postgres.NewDemandRepository(pool), &clients.Inventory{Client: client.New(cfg.InternalKey, cfg.URLs)})
	httpadapter.Mount(mux, s)
	server.Health(mux, "demand", pool)
	return server.Application{Handler: httpx.Internal(cfg.InternalKey, mux), Close: pool.Close}, nil
}
