// Package notificationservice is the composition root of the notification service.
package notificationservice

import (
	"context"
	"net/http"

	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/httpx"
	database "github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/internal/platform/server"
	httpadapter "github.com/depsi4/cargos/notification_service/internal/adapters/primary/http"
	"github.com/depsi4/cargos/notification_service/internal/adapters/secondary/postgres"
	"github.com/depsi4/cargos/notification_service/internal/core/services"
	"github.com/depsi4/cargos/notification_service/migrations"
)

func New(ctx context.Context, cfg config.Config) (server.Application, error) {
	pool, err := database.Open(ctx, cfg.DatabaseURL, "notification", migrations.Files)
	if err != nil {
		return server.Application{}, err
	}
	mux := http.NewServeMux()
	s := services.New(postgres.NewNotificationRepository(pool))
	httpadapter.Mount(mux, s)
	server.Health(mux, "notification", pool)
	return server.Application{Handler: httpx.Internal(cfg.InternalKey, mux), Close: pool.Close}, nil
}
