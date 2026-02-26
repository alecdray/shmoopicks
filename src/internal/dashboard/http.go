package dashboard

import (
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/feed"
	"shmoopicks/src/internal/musicbrainz"
	"shmoopicks/src/internal/spotify"
)

type HttpHandler struct {
	spotifyAuth *spotify.AuthService
	mb          *musicbrainz.Service
	feedService *feed.Service
}

func NewHttpHandler(spotifyAuth *spotify.AuthService, mb *musicbrainz.Service, feedService *feed.Service) *HttpHandler {
	return &HttpHandler{
		spotifyAuth: spotifyAuth,
		mb:          mb,
		feedService: feedService,
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

	for _, feed := range feeds {
		if feed.Kind == models.FeedKindSpotify && feed.IsStale() {
			err := h.feedService.SyncSpotifyFeed(ctx, feed)
			if err != nil {
				slog.Error("failed to sync spotify feed", "error", err)
			}
		}
	}

	dashboardPage := DashboardPage(DashboardPageProps{
		Feeds: feeds,
	})
	dashboardPage.Render(r.Context(), w)
}
