package main

import (
	"log/slog"
	"os"

	"github.com/depsi4/cargos/demand_service"
	"github.com/depsi4/cargos/internal/platform/server"
)

func main() {
	if err := server.Run("demand", demandservice.New); err != nil {
		slog.Error("service stopped", "service", "demand", "error", err)
		os.Exit(1)
	}
}
