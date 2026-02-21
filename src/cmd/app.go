package main

import (
	"context"
	"log/slog"
	"shmoopicks/src/internal/core/config"
	"shmoopicks/src/internal/server"
)

func main() {
	slog.Info("Starting app")
	ctx := context.Background()

	config := config.LoadConfig()

	if err := config.ValidateConfig(); err != nil {
		slog.Error("Error validating config", "error", err)
		return
	}

	server.Start(ctx, config)
}
