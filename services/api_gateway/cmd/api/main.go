package main

import (
	"log/slog"
	"os"

	gateway "github.com/depsi4/cargos/api_gateway"
	"github.com/depsi4/cargos/internal/platform/server"
)

func main() {
	if err := server.Run("gateway", gateway.New); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}
