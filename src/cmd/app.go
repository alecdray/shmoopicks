package main

import (
	"context"
	"log/slog"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/server"
)

func main() {
	slog.Info("Starting app")
	ctx := context.Background()

	config := app.LoadConfig()

	if err := config.ValidateConfig(); err != nil {
		slog.Error("Error validating config", "error", err)
		return
	}

	server.Start(ctx, app.NewApp(*config))
}
