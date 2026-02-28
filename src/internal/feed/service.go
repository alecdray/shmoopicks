package feed

import (
	"context"
	"fmt"
	"log/slog"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/core/sqlx"
	"shmoopicks/src/internal/core/timex"
	"shmoopicks/src/internal/core/utils"
	"shmoopicks/src/internal/library"
	"shmoopicks/src/internal/spotify"
	"time"

	"github.com/google/uuid"
)

const (
	MinStaleDuration = 1 * timex.Day
)

type FeedDTO struct {
	ID                  string
	UserID              string
	Kind                models.FeedKind
	LastSyncStatus      models.FeedSyncStatus
	LastSyncCompletedAt *time.Time
	LastSyncStartedAt   *time.Time
}

func NewFeedDTOFromModel(model sqlc.Feed) *FeedDTO {
	dto := &FeedDTO{
		ID:             model.ID,
		UserID:         model.UserID,
		Kind:           model.Kind,
		LastSyncStatus: model.LastSyncStatus,
	}

	if model.LastSyncStartedAt.Valid {
		dto.LastSyncStartedAt = &model.LastSyncStartedAt.Time
	}

	if model.LastSyncCompletedAt.Valid {
		dto.LastSyncCompletedAt = &model.LastSyncCompletedAt.Time
	}

	return dto
}

func (f FeedDTO) IsUnsyned() bool {
	return f.LastSyncStatus == models.FeedSyncStatusNone
}

func (f FeedDTO) IsSyncing() bool {
	return f.LastSyncStatus == models.FeedSyncStatusPending
}

func (f FeedDTO) IsSynced() bool {
	return f.LastSyncStatus == models.FeedSyncStatusSuccess
}

func (f FeedDTO) IsStale() bool {
	if f.IsUnsyned() {
		return false
	}
	minStaleTime := time.Now().Add(-MinStaleDuration)
	return f.LastSyncCompletedAt.Before(minStaleTime)
}

func (f *FeedDTO) SetSyncFailed() {
	f.LastSyncStatus = models.FeedSyncStatusFailure
	f.LastSyncCompletedAt = utils.NewPointer(time.Now())
}

func (f *FeedDTO) SetSyncSuccess() {
	f.LastSyncStatus = models.FeedSyncStatusSuccess
	f.LastSyncCompletedAt = utils.NewPointer(time.Now())
}

func (f *FeedDTO) SetSyncing() {
	f.LastSyncStatus = models.FeedSyncStatusPending
	f.LastSyncStartedAt = utils.NewPointer(time.Now())
}

type Service struct {
	db             *db.DB
	spotifyService *spotify.Service
	libraryService *library.Service
}

func NewService(db *db.DB, spotifyService *spotify.Service, libraryService *library.Service) *Service {
	return &Service{
		db:             db,
		spotifyService: spotifyService,
		libraryService: libraryService,
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

func (s *Service) UpdateFeed(ctx context.Context, feed FeedDTO) (*FeedDTO, error) {
	feedModel, err := s.db.Queries().UpdateFeed(ctx, sqlc.UpdateFeedParams{
		ID:                  feed.ID,
		LastSyncStatus:      feed.LastSyncStatus,
		LastSyncStartedAt:   sqlx.NewNullTime(feed.LastSyncStartedAt),
		LastSyncCompletedAt: sqlx.NewNullTime(feed.LastSyncCompletedAt),
	})
	if err != nil {
		return nil, err
	}
	return NewFeedDTOFromModel(feedModel), nil
}

func (s *Service) syncSpotifyFeed(ctx contextx.ContextX, feed FeedDTO) error {
	savedAlbums, err := s.spotifyService.GetUsersSavedAlbums(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get user saved albums: %w", err)
		return err
	}

	albumsToSync := make([]library.AlbumDTO, len(savedAlbums))
	for i, album := range savedAlbums {
		var addedAt *time.Time = nil
		_addedAt, err := time.Parse(time.RFC3339, album.AddedAt)
		if err != nil {
			slog.Error("failed to parse added at time during syncSpotifyFeed", err)
		} else {
			addedAt = &_addedAt
		}

		lib := library.AlbumDTO{
			ID:        uuid.NewString(),
			SpotifyID: album.ID.String(),
			Title:     album.Name,
			Artists:   make([]library.ArtistDTO, len(album.Artists)),
			Tracks:    []library.TrackDTO{},
			Releases: []library.ReleaseDTO{
				{
					ID:      uuid.NewString(),
					Format:  models.ReleaseFormatDigital,
					AddedAt: addedAt,
				},
			},
		}

		for i, artist := range album.Artists {
			lib.Artists[i] = library.ArtistDTO{
				ID:        uuid.NewString(),
				SpotifyID: artist.ID.String(),
				Name:      artist.Name,
			}
		}

		for _, track := range album.Tracks.Tracks {
			lib.Tracks = append(lib.Tracks, library.TrackDTO{
				ID:        uuid.NewString(),
				SpotifyID: track.ID.String(),
				Title:     track.Name,
			})
		}

		albumsToSync[i] = lib
	}

	err = s.libraryService.AddAlbumsToLibrary(ctx, feed.UserID, albumsToSync)
	if err != nil {
		err = fmt.Errorf("failed to add albums to library: %w", err)
		return err
	}

	return nil
}

func (s *Service) SyncSpotifyFeed(ctx contextx.ContextX, feed FeedDTO) (*FeedDTO, error) {
	if feed.Kind != models.FeedKindSpotify {
		return nil, fmt.Errorf("feed kind must be spotify")
	}

	feed.SetSyncing()
	_, err := s.UpdateFeed(ctx, feed)
	if err != nil {
		err = fmt.Errorf("failed to update feed on sync start: %w", err)
		return nil, err
	}

	err = s.syncSpotifyFeed(ctx, feed)
	if err != nil {
		err = fmt.Errorf("failed to sync spotify feed: %w", err)

		feed.SetSyncFailed()
		_, err := s.UpdateFeed(ctx, feed)
		if err != nil {
			slog.Error("failed to update feed on sync error", "error", err)
		}

		return nil, err
	}

	feed.SetSyncSuccess()
	_, err = s.UpdateFeed(ctx, feed)
	if err != nil {
		err = fmt.Errorf("failed to update feed on sync success: %w", err)
		return nil, err
	}

	return &feed, nil
}
