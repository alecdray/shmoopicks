package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"shmoopicks/src/internal/core/apphttp"
	"shmoopicks/src/internal/core/config"
	"shmoopicks/src/internal/dashboard"
)

func Start(ctx context.Context, config *config.Config) {
	rootMux := apphttp.NewWrappedMux(*config)

	rootMux.Handle("/static/", apphttp.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/public")))))

	dashboardHandler := dashboard.NewHttpHandler()
	rootMux.Handle("/dashboard", apphttp.AppHandlerFunc(dashboardHandler.HandleGetDashboard))

	addr := fmt.Sprintf(":%s", config.Port)
	slog.Info("Starting server", "addr", addr)
	err := http.ListenAndServe(addr, rootMux)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
