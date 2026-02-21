package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"shmoopicks/src/internal/auth"
	"shmoopicks/src/internal/core/apphttp"
	"shmoopicks/src/internal/core/config"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/dashboard"
)

func Start(ctx context.Context, config *config.Config) {
	db, err := db.NewDB(config.DbPath)
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		os.Exit(1)
	}

	rootMux := apphttp.NewWrappedMux(*config)

	rootMux.Handle("/static/", apphttp.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/public")))))

	authHandler := auth.NewHttpHandler(db)
	rootMux.Handle("/", apphttp.AppHandlerFunc(authHandler.GetLoginPage))

	appMux := apphttp.NewWrappedMux(*config, apphttp.JwtMiddleware)
	rootMux.Use("/app/", appMux)

	dashboardHandler := dashboard.NewHttpHandler()
	appMux.Handle("/app/dashboard", apphttp.AppHandlerFunc(dashboardHandler.GetDashboardPage))

	addr := fmt.Sprintf(":%s", config.Port)
	slog.Info("Starting server", "addr", addr)
	err = http.ListenAndServe(addr, rootMux)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
