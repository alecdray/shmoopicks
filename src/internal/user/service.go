package user

import (
	"context"
	"errors"
	"fmt"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/cryptox"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/core/sqlx"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID                  string
	SpotifyID           string
	spotifyRefreshToken *string
}

func NewUserDTOFromModel(model sqlc.User) *UserDTO {
	user := &UserDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
	}

	if model.SpotifyRefreshToken.Valid {
		user.spotifyRefreshToken = &model.SpotifyRefreshToken.String
	}

	return user
}

func (u *UserDTO) SpotifyRefreshToken(secret string) *string {
	if u.spotifyRefreshToken == nil {
		return nil
	}

	decrypted, err := cryptox.SymmetricDecrypt(*u.spotifyRefreshToken, secret)
	if err != nil {
		return nil
	}

	return &decrypted
}

type Service struct {
	db *db.DB
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

func (s *Service) UpsertSpotifyUser(ctx contextx.ContextX, spotifyId string, spotifyRefreshToken string) (*UserDTO, error) {
	app, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		return nil, err
	}

	encryptedSpotifyRefreshToken, err := cryptox.SymmetricEncrypt(spotifyRefreshToken, app.Config().SpotifyTokenSecret)
	if err != nil {
		err = fmt.Errorf("failed to encrypt spotify refresh token: %w", err)
		return nil, err
	}

	user, err := s.db.Queries().UpsertSpotifyUser(ctx, sqlc.UpsertSpotifyUserParams{
		ID:                  uuid.New().String(),
		SpotifyID:           spotifyId,
		SpotifyRefreshToken: sqlx.NewNullString(encryptedSpotifyRefreshToken),
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

	if userId == "" {
		app, err := ctx.App()
		if err != nil {
			err = fmt.Errorf("failed to get app: %w", err)
			return nil, err
		}

		userId = *app.Claims().UserID
	}

	if userId != "" {
		userDto, err := s.GetUserById(ctx, userId)
		if err != nil {
			err = fmt.Errorf("failed to get user by id: %w", err)
			return nil, err
		}
		return userDto, nil
	}

	return nil, errors.New("unauthorized")
}
