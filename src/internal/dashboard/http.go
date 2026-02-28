package dashboard

import (
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/feed"
	"shmoopicks/src/internal/library"
	"shmoopicks/src/internal/musicbrainz"
	"shmoopicks/src/internal/spotify"
)

type HttpHandler struct {
	spotifyAuth    *spotify.AuthService
	mb             *musicbrainz.Service
	feedService    *feed.Service
	libraryService *library.Service
}

func NewHttpHandler(spotifyAuth *spotify.AuthService, mb *musicbrainz.Service, feedService *feed.Service, libraryService *library.Service) *HttpHandler {
	return &HttpHandler{
		spotifyAuth:    spotifyAuth,
		mb:             mb,
		feedService:    feedService,
		libraryService: libraryService,
	}
}

func (h *HttpHandler) GetDashboardPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feeds, err := h.feedService.GetUsersFeeds(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range feeds {
		if f.Kind == models.FeedKindSpotify && !f.IsSynced() {
			go func(feed feed.FeedDTO) {
				_, err := h.feedService.SyncSpotifyFeed(ctx, feed)
				if err != nil {
					slog.Error("failed to sync spotify feed", "error", err)
				}
			}(f)
		}
	}

	library, err := h.libraryService.GetLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get library: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	library.Albums.SortByDate(false)

	dashboardPage := DashboardPage(DashboardPageProps{
		Library: library,
		Feeds:   feeds,
	})
	dashboardPage.Render(r.Context(), w)
}

func (h *HttpHandler) GetAlbumsTable(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	library, err := h.libraryService.GetLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get library: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if library == nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	albums := library.Albums
	sortBy := r.URL.Query().Get("sortBy")
	dir := r.URL.Query().Get("dir")

	// Default to ascending if not specified
	ascending := dir != "desc"

	// Sort albums based on sortBy parameter
	switch sortBy {
	case "album":
		albums.SortByTitle(ascending)
	case "artist":
		albums.SortByArtist(ascending)
	case "date":
		albums.SortByDate(ascending)
	}

	component := AlbumsTable(albums, sortBy, dir)
	component.Render(r.Context(), w)
}

func (h *HttpHandler) GetFeedsDropdown(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feeds, err := h.feedService.GetUsersFeeds(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render content first
	contentComponent := feedsDropdownContent(feeds)
	contentComponent.Render(r.Context(), w)

	// Render button as OOB swap
	buttonComponent := feedsDropdownButton(feeds, true)
	buttonComponent.Render(r.Context(), w)
}
