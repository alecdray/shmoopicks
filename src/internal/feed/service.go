package feed

import (
	"context"
	"log/slog"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/core/timex"
	"shmoopicks/src/internal/spotify"
	"time"

	"github.com/google/uuid"
)

const (
	MinStaleDuration = 1 * timex.Day
)

type FeedDTO struct {
	ID           string
	UserID       string
	Kind         models.FeedKind
	LastSyncedAt *time.Time
}

func NewFeedDTOFromModel(model sqlc.Feed) *FeedDTO {
	return &FeedDTO{
		ID:     model.ID,
		UserID: model.UserID,
		Kind:   model.Kind,
	}
}

func (f FeedDTO) IsUnsyned() bool {
	return f.LastSyncedAt == nil
}

func (f FeedDTO) IsStale() bool {
	if f.LastSyncedAt == nil {
		return true
	}
	minStaleTime := time.Now().Add(-MinStaleDuration)
	return f.LastSyncedAt.Before(minStaleTime)
}

type Service struct {
	db             *db.DB
	spotifyService *spotify.Service
}

func NewService(db *db.DB, spotifyService *spotify.Service) *Service {
	return &Service{
		db:             db,
		spotifyService: spotifyService,
	}
}

func (s *Service) UpsertFeed(ctx context.Context, userID string, kind models.FeedKind) (*FeedDTO, error) {
	feed, err := s.db.Queries().UpsertFeed(ctx, sqlc.UpsertFeedParams{
		ID:     uuid.New().String(),
		UserID: userID,
		Kind:   kind,
	})
	if err != nil {
		return nil, err
	}
	return NewFeedDTOFromModel(feed), nil
}

func (s *Service) GetUsersFeeds(ctx context.Context, userID string) ([]FeedDTO, error) {
	feeds, err := s.db.Queries().GetFeedsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	var feedDTOs []FeedDTO
	for _, feed := range feeds {
		feedDTOs = append(feedDTOs, *NewFeedDTOFromModel(feed))
	}

	return feedDTOs, nil
}

func (s *Service) SyncSpotifyFeed(ctx contextx.ContextX, feed FeedDTO) error {
	albums, err := s.spotifyService.GetUsersSavedAlbums(ctx)
	if err != nil {
		return err
	}

	for _, album := range albums {
		slog.Info("user album", "album", album.Name)
	}

	tracks, err := s.spotifyService.GetUsersSavedTracks(ctx)
	if err != nil {
		return err
	}

	for _, track := range tracks {
		slog.Info("user track", "track", track.Name)
	}

	return nil
}
