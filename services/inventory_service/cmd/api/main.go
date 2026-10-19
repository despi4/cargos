package main

import (
	"log/slog"
	"os"

	"github.com/depsi4/cargos/internal/platform/server"
	"github.com/depsi4/cargos/inventory_service"
)

func main() {
	if err := server.Run("inventory", inventoryservice.New); err != nil {
		slog.Error("service stopped", "service", "inventory", "error", err)
		os.Exit(1)
	}
}
