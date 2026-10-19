package main

import (
	"log/slog"
	"os"

	"github.com/depsi4/cargos/internal/platform/server"
	"github.com/depsi4/cargos/notification_service"
)

func main() {
	if err := server.Run("notification", notificationservice.New); err != nil {
		slog.Error("service stopped", "service", "notification", "error", err)
		os.Exit(1)
	}
}
