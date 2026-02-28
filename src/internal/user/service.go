package user

import (
	"context"
	"errors"
	"fmt"
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

func (s *Service) GetUserById(ctx context.Context, id string) (*UserDTO, error) {
	user, err := s.db.Queries().GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewUserDTOFromModel(user), nil
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

func (s *Service) GetUserFromCtx(ctx contextx.ContextX) (*UserDTO, error) {
	userId, err := ctx.UserId()
	if errors.Is(err, contextx.ErrEmptyValue) {
		userId = ""
	} else if err != nil {
		err = fmt.Errorf("failed to get user id: %w", err)
		return nil, err
	}

	if userId != "" {
		userDto, err := s.GetUserById(ctx, userId)
		if err != nil {
			err = fmt.Errorf("failed to get user by id: %w", err)
			return nil, err
		}
		return userDto, nil
	}

	app, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		return nil, err
	}

	if app.Claims() != nil && app.Claims().SpotifyToken != nil {
		user, err := s.spotifyService.GetUser(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get spotify user: %w", err)
			return nil, err
		}

		userDTO, err := s.GetUserBySpotifyID(ctx, user.ID)
		if err != nil {
			err = fmt.Errorf("failed to get user by spotify id: %w", err)
			return nil, err
		}
		return userDTO, nil
	}

	return nil, errors.New("unauthorized")
}
