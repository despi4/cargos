package main

import (
	"log/slog"
	"os"

	"github.com/depsi4/cargos/internal/platform/server"
	"github.com/depsi4/cargos/prediction_service"
)

func main() {
	if err := server.Run("prediction", predictionservice.New); err != nil {
		slog.Error("service stopped", "service", "prediction", "error", err)
		os.Exit(1)
	}
}
