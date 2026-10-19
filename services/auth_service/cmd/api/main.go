package main

import (
	"log/slog"
	"os"

	"github.com/depsi4/cargos/auth_service"
	"github.com/depsi4/cargos/internal/platform/server"
)

func main() {
	if err := server.Run("auth", authservice.New); err != nil {
		slog.Error("service stopped", "service", "auth", "error", err)
		os.Exit(1)
	}
}
