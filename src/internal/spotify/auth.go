package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"time"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	ErrFailedToGetToken = errors.New("failed to get token")
	ErrStateMismatch    = errors.New("state mismatch")
)

type AuthService struct {
	*spotifyauth.Authenticator
}

func NewAuthService(clientID, clientSecret, redirectURI string, scopes ...string) *AuthService {
	return &AuthService{
		spotifyauth.New(
			spotifyauth.WithClientID(clientID),
			spotifyauth.WithClientSecret(clientSecret),
			spotifyauth.WithRedirectURL(redirectURI),
			spotifyauth.WithScopes(scopes...),
		),
	}
}

func (auth *AuthService) GetClientFromCallback(ctx contextx.ContextX, state string, r *http.Request) (*spotify.Client, error) {
	token, err := auth.Token(ctx, state, r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}
	if st := r.FormValue("state"); st != state {
		return nil, ErrStateMismatch
	}

	return spotify.New(auth.Client(ctx, token)), nil
}

func (auth *AuthService) GetClientFromRefreshToken(ctx context.Context, refreshToken string) (*spotify.Client, error) {
	partialToken := oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second),
	}

	token, err := auth.RefreshToken(ctx, &partialToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}

	return spotify.New(auth.Client(ctx, token)), nil
}
