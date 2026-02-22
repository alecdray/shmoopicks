package spotify

import (
	"errors"
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/appctx"

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

func (auth *AuthService) GetClientWithCallback(ctx appctx.Ctx, state string, r *http.Request) (*spotify.Client, error) {
	token, err := auth.Token(ctx, state, r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}
	if st := r.FormValue("state"); st != state {
		return nil, ErrStateMismatch
	}

	return spotify.New(auth.Client(ctx, token)), nil
}

func (auth *AuthService) GetClient(ctx appctx.Ctx, token *oauth2.Token) *spotify.Client {
	return spotify.New(auth.Client(ctx, token))
}
