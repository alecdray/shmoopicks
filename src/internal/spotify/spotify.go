package spotify

import (
	"shmoopicks/src/internal/core/contextx"
	"time"

	spotify "github.com/zmb3/spotify/v2"
)

type Service struct {
	*spotify.Client
}

func NewService(client *spotify.Client) *Service {
	return &Service{
		Client: client,
	}
}

func (s *Service) GetRecentlySavedTracks(ctx contextx.ContextX, window time.Duration) ([]spotify.SavedTrack, error) {
	var userTracks []spotify.SavedTrack = nil
	minTime := time.Now().Add(-window)
	maxTime := time.Now()
	offset := 0
	for userTracks == nil || maxTime.After(minTime) {
		tracks, err := s.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset(offset))
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
	var recentlyPlayedTracks []spotify.RecentlyPlayedItem = nil
	minTime := time.Now().Add(-window)
	maxTime := time.Now()
	for recentlyPlayedTracks == nil || maxTime.After(minTime) {
		tracks, err := s.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{
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
