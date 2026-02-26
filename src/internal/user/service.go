package user

import (
	"context"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/spotify"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID        string
	SpotifyID string
}

func NewUserDTOFromModel(user sqlc.User) *UserDTO {
	return &UserDTO{
		ID:        user.ID,
		SpotifyID: user.SpotifyID,
	}
}

type Service struct {
	db             *db.DB
	spotifyService *spotify.Service
}

func NewService(db *db.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) GetUserBySpotifyID(ctx context.Context, spotifyId string) (*UserDTO, error) {
	user, err := s.db.Queries().GetUserBySpotifyId(ctx, spotifyId)
	if err != nil {
		return nil, err
	}
	return NewUserDTOFromModel(user), nil
}

func (s *Service) UpsertSpotifyUser(ctx context.Context, spotifyId string) (*UserDTO, error) {
	user, err := s.db.Queries().UpsertSpotifyUser(ctx, sqlc.UpsertSpotifyUserParams{
		ID:        uuid.New().String(),
		SpotifyID: spotifyId,
	})
	if err != nil {
		return nil, err
	}
	return NewUserDTOFromModel(user), nil
}

func (s *Service) LoginSpotifyUser(ctx contextx.ContextX) (*UserDTO, error) {
	user, err := s.spotifyService.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	userDTO, err := s.GetUserBySpotifyID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return userDTO, nil
}
