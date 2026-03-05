package review

import (
	"context"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/core/sqlx"

	"github.com/google/uuid"
)

type AlbumRatingDTO struct {
	ID      string
	UserID  string
	AlbumID string
	Rating  *float64
}

func NewAlbumRatingDTOFromModel(model sqlc.AlbumRating) *AlbumRatingDTO {
	dto := &AlbumRatingDTO{
		ID:      model.ID,
		UserID:  model.UserID,
		AlbumID: model.AlbumID,
	}

	if model.Rating.Valid {
		dto.Rating = &model.Rating.Float64
	}

	return dto
}

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) UpdateRating(ctx context.Context, userId, albumId string, rating float64) (*AlbumRatingDTO, error) {
	model, err := s.db.Queries().UpsertAlbumRating(ctx, sqlc.UpsertAlbumRatingParams{
		ID:      uuid.NewString(),
		UserID:  userId,
		AlbumID: albumId,
		Rating:  sqlx.NewNullFloat64(rating),
	})
	if err != nil {
		return nil, err
	}

	return NewAlbumRatingDTOFromModel(model), nil
}
