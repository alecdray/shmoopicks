package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"shmoopicks/src/internal/auth"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/httpx"
	"shmoopicks/src/internal/dashboard"
	"shmoopicks/src/internal/feed"
	"shmoopicks/src/internal/library"
	"shmoopicks/src/internal/musicbrainz"
	"shmoopicks/src/internal/spotify"
	"shmoopicks/src/internal/user"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func Start(ctx context.Context, app app.App) {
	db, err := db.NewDB(app.Config().DbPath)
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	mbClient, err := musicbrainz.NewClient(
		app.Config().AppName,
		app.Config().AppVersion,
		musicbrainz.WithContactEmail(app.Config().ContactEmail),
	)
	if err != nil {
		slog.Error("Failed to create MusicBrainz client", "error", err)
		os.Exit(1)
	}

	userService := user.NewService(db)

	mbService := musicbrainz.NewService(mbClient)

	spotifyAuthService := spotify.NewAuthService(
		app.Config().SpotifyClientId,
		app.Config().SpotifyClientSecret,
		fmt.Sprintf("%s/spotify/callback", app.Config().Host),
		spotifyauth.ScopeUserLibraryRead,
		spotifyauth.ScopeUserReadRecentlyPlayed,
	)

	spotifyService := spotify.NewService()

	libraryService := library.NewService(db)

	feedService := feed.NewService(db, spotifyService, libraryService)

	rootMux := httpx.NewMux(app, httpx.RequestLoggingMiddleware)

	rootMux.Handle("/static/", httpx.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/public")))))

	authHandler := auth.NewHttpHandler(db, spotifyAuthService, userService, feedService)
	rootMux.Handle("/{$}", httpx.HandlerFunc(authHandler.GetLoginPage))
	rootMux.Handle("/logout", httpx.HandlerFunc(authHandler.Logout))
	rootMux.Handle("/spotify/callback", httpx.HandlerFunc(authHandler.AuthorizeSpotify))

	appMux := httpx.NewMux(app, httpx.JwtMiddleware(spotifyService, userService))
	rootMux.Use("/app/", appMux)

	dashboardHandler := dashboard.NewHttpHandler(spotifyAuthService, mbService, feedService, libraryService)
	appMux.Handle("/app/dashboard", httpx.HandlerFunc(dashboardHandler.GetDashboardPage))

	// Not found handler, must be registered after all other handlers
	rootMux.HandleFunc("/not-found", httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}))

	addr := fmt.Sprintf(":%s", app.Config().Port)
	slog.Info("Starting server", "addr", addr)
	err = http.ListenAndServe(addr, rootMux)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
