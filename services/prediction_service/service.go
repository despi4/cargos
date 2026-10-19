// Package predictionservice is the composition root of the prediction service.
package predictionservice

import (
	"context"
	"net/http"

	client "github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/httpx"
	database "github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/internal/platform/server"
	httpadapter "github.com/depsi4/cargos/prediction_service/internal/adapters/primary/http"
	clients "github.com/depsi4/cargos/prediction_service/internal/adapters/secondary/httpclient"
	"github.com/depsi4/cargos/prediction_service/internal/adapters/secondary/postgres"
	"github.com/depsi4/cargos/prediction_service/internal/core/services"
	"github.com/depsi4/cargos/prediction_service/migrations"
)

func New(ctx context.Context, cfg config.Config) (server.Application, error) {
	pool, err := database.Open(ctx, cfg.DatabaseURL, "prediction", migrations.Files)
	if err != nil {
		return server.Application{}, err
	}
	mux := http.NewServeMux()
	transport := client.New(cfg.InternalKey, cfg.URLs)
	s := services.New(postgres.NewPredictionRepository(pool), &clients.Sources{Client: transport}, &clients.Notifier{Client: transport})
	httpadapter.Mount(mux, s)
	server.Health(mux, "prediction", pool)
	return server.Application{Handler: httpx.Internal(cfg.InternalKey, mux), Close: pool.Close, Tick: s.DeliverPending}, nil
}
