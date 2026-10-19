package gateway

import (
	"context"
	"net/http"

	httpadapter "github.com/depsi4/cargos/api_gateway/internal/adapters/primary/http"
	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/server"
)

func New(_ context.Context, cfg config.Config) (server.Application, error) {
	mux := http.NewServeMux()
	server.Health(mux, "gateway", nil)
	return server.Application{Handler: httpadapter.NewRouter(mux, cfg.URLs, cfg.InternalKey)}, nil
}
