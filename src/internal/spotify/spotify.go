package spotify

import (
	"fmt"
	"shmoopicks/src/internal/core/contextx"
	"time"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Client(ctx contextx.ContextX) (*spotify.Client, error) {
	a, err := ctx.App()
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	claims := a.Claims()
	if claims == nil || claims.SpotifyToken == nil {
		return nil, fmt.Errorf("spotify token not found in JWT claims")
	}

	return spotify.New(spotifyauth.New().Client(ctx, claims.SpotifyToken)), nil
}

func (s *Service) GetUser(ctx contextx.ContextX) (*spotify.PrivateUser, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	return client.CurrentUser(ctx)
}

func (s *Service) GetRecentlySavedTracks(ctx contextx.ContextX, window time.Duration) ([]spotify.SavedTrack, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	var userTracks []spotify.SavedTrack = nil
	minTime := time.Now().Add(-window)
	maxTime := time.Now()
	offset := 0
	for userTracks == nil || maxTime.After(minTime) {
		tracks, err := client.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		offset += len(tracks.Tracks)

		if len(tracks.Tracks) == 0 {
			break
		} else if userTracks == nil {
			userTracks = make([]spotify.SavedTrack, 0, len(tracks.Tracks))
		}

		for _, track := range tracks.Tracks {
			addedAt, err := time.Parse(time.RFC3339, track.AddedAt)
			if err != nil {
				return nil, err
			}

			if addedAt.After(minTime) {
				userTracks = append(userTracks, track)
			}

			if addedAt.Before(maxTime) {
				maxTime = addedAt
			}
		}
	}

	return userTracks, nil
}

func (s *Service) GetRecentlyPlayedTracks(ctx contextx.ContextX, window time.Duration) ([]spotify.RecentlyPlayedItem, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	var recentlyPlayedTracks []spotify.RecentlyPlayedItem = nil
	minTime := time.Now().Add(-window)
	maxTime := time.Now()
	for recentlyPlayedTracks == nil || maxTime.After(minTime) {
		tracks, err := client.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{
			Limit:         50,
			BeforeEpochMs: maxTime.UnixMilli(),
		})
		if err != nil {
			return nil, err
		}

		if len(tracks) == 0 {
			break
		} else if recentlyPlayedTracks == nil {
			recentlyPlayedTracks = make([]spotify.RecentlyPlayedItem, 0, len(tracks))
		}

		for _, track := range tracks {
			if track.PlayedAt.After(minTime) {
				recentlyPlayedTracks = append(recentlyPlayedTracks, track)
			}

			if track.PlayedAt.Before(maxTime) {
				maxTime = track.PlayedAt
			}
		}
	}
	return recentlyPlayedTracks, nil
}

func (s *Service) GetUsersSavedAlbums(ctx contextx.ContextX) ([]spotify.SavedAlbum, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	var collectedAlbums []spotify.SavedAlbum = make([]spotify.SavedAlbum, 0)
	limit := 50
	offset := 0
	for offset < 1_000 {
		albums, err := client.CurrentUsersAlbums(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		if len(albums.Albums) == 0 {
			break
		}

		collectedAlbums = append(collectedAlbums, albums.Albums...)

		offset += len(albums.Albums)
	}
	return collectedAlbums, nil
}

func (s *Service) GetUsersSavedTracks(ctx contextx.ContextX) ([]spotify.SavedTrack, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	var collectedTracks []spotify.SavedTrack = make([]spotify.SavedTrack, 0)
	limit := 50
	offset := 0
	for offset < 1_000 {
		tracks, err := client.CurrentUsersTracks(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		if len(tracks.Tracks) == 0 {
			break
		}

		collectedTracks = append(collectedTracks, tracks.Tracks...)

		offset += len(tracks.Tracks)
	}
	return collectedTracks, nil
}
