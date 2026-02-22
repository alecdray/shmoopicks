package spotify

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/contextx"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
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

func (auth *AuthService) GetClientWithCallback(ctx contextx.ContextX, state string, r *http.Request) (*spotify.Client, error) {
	token, err := auth.Token(ctx, state, r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}
	if st := r.FormValue("state"); st != state {
		return nil, ErrStateMismatch
	}

	return spotify.New(auth.Client(ctx, token)), nil
}

func (auth *AuthService) GetClient(ctx contextx.ContextX) (*spotify.Client, error) {
	a, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		slog.DebugContext(ctx, err.Error())
	}

	claims, err := a.GetClaims()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT claims: %w", err)
	}

	if claims.SpotifyToken == nil {
		return nil, fmt.Errorf("spotify token not found in JWT claims")
	}

	return spotify.New(auth.Client(ctx, claims.SpotifyToken)), nil
}
