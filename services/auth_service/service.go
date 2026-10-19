// Package authservice is the composition root of the auth service.
package authservice

import (
	"context"
	"net/http"

	httpadapter "github.com/depsi4/cargos/auth_service/internal/adapters/primary/http"
	"github.com/depsi4/cargos/auth_service/internal/adapters/secondary/postgres"
	"github.com/depsi4/cargos/auth_service/internal/adapters/secondary/security"
	"github.com/depsi4/cargos/auth_service/internal/core/services"
	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/server"
)

func New(ctx context.Context, cfg config.Config) (server.Application, error) {
	db, err := postgres.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return server.Application{}, err
	}
	pool := db.Pool
	mux := http.NewServeMux()
	s := services.New(postgres.NewUserRepo(pool), postgres.NewSessionRepository(pool), security.NewTokens(cfg.JWTSecret), security.PasswordHasher{})
	httpadapter.Mount(mux, s)
	server.Health(mux, "auth", pool)
	return server.Application{Handler: mux, Close: pool.Close}, nil
}
