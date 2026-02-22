package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"shmoopicks/src/internal/auth"
	"shmoopicks/src/internal/core/appctx"
	"shmoopicks/src/internal/core/apphttp"
	"shmoopicks/src/internal/core/config"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/dashboard"
	"shmoopicks/src/internal/spotify"
)

func Start(ctx context.Context, config *config.Config) {
	db, err := db.NewDB(config.DbPath)
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		os.Exit(1)
	}

	rootMux := apphttp.NewMux(*config, apphttp.RequestLoggingMiddleware)

	rootMux.Handle("/static/", apphttp.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/public")))))

	spotifyAuthService := spotify.NewAuthService(
		config.SpotifyClientId,
		config.SpotifyClientSecret,
		fmt.Sprintf("%s/spotify/callback", config.Host),
	)
	authHandler := auth.NewHttpHandler(db, spotifyAuthService)
	rootMux.Handle("/{$}", apphttp.HandlerFunc(authHandler.GetLoginPage))
	rootMux.Handle("/logout", apphttp.HandlerFunc(authHandler.Logout))
	rootMux.Handle("/spotify/callback", apphttp.HandlerFunc(authHandler.AuthorizeSpotify))

	appMux := apphttp.NewMux(*config, apphttp.JwtMiddleware)
	rootMux.Use("/app/", appMux)

	dashboardHandler := dashboard.NewHttpHandler()
	appMux.Handle("/app/dashboard", apphttp.HandlerFunc(dashboardHandler.GetDashboardPage))

	// Not found handler, must be registered after all other handlers
	rootMux.HandleFunc("/not-found", apphttp.HandlerFunc(func(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}))

	addr := fmt.Sprintf(":%s", config.Port)
	slog.Info("Starting server", "addr", addr)
	err = http.ListenAndServe(addr, rootMux)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
