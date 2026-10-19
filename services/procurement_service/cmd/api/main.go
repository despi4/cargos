package main

import (
	"log/slog"
	"os"

	"github.com/depsi4/cargos/internal/platform/server"
	"github.com/depsi4/cargos/procurement_service"
)

func main() {
	if err := server.Run("procurement", procurementservice.New); err != nil {
		slog.Error("service stopped", "service", "procurement", "error", err)
		os.Exit(1)
	}
}
