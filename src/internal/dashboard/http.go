package dashboard

import (
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/spotify"
	"time"
)

type HttpHandler struct {
	spotifyAuth *spotify.AuthService
}

func NewHttpHandler(spotifyAuth *spotify.AuthService) *HttpHandler {
	return &HttpHandler{
		spotifyAuth: spotifyAuth,
	}
}

func (h *HttpHandler) GetDashboardPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	client, err := h.spotifyAuth.GetClient(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get spotify client: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	spotifyService := spotify.NewService(client)

	window := time.Hour * 24 * 5

	savedTracks, err := spotifyService.GetRecentlySavedTracks(ctx, window)
	if err != nil {
		err = fmt.Errorf("failed to get recently saved tracks: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	playedTracks, err := spotifyService.GetRecentlyPlayedTracks(ctx, window)
	if err != nil {
		err = fmt.Errorf("failed to get recently played tracks: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	playedTrackIdsSet := make(map[string]struct{})
	for _, track := range playedTracks {
		playedTrackIdsSet[string(track.Track.ID)] = struct{}{}
	}

	dashboardPage := DashboardPage(DashboardPageProps{
		Tracks:               savedTracks,
		RecentlyPlayedTracks: playedTrackIdsSet,
	})
	dashboardPage.Render(r.Context(), w)
}
