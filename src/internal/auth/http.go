package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/httpx"
	"shmoopicks/src/internal/spotify"
)

type HttpHandler struct {
	db          *db.DB
	spotifyAuth *spotify.AuthService
}

func NewHttpHandler(db *db.DB, spotifyAuth *spotify.AuthService) *HttpHandler {
	return &HttpHandler{
		db:          db,
		spotifyAuth: spotifyAuth,
	}
}

func (h *HttpHandler) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		httpx.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	claims, err := app.ValidateClaimsFromRequest(r, a.Config().JwtSecret)
	if err != nil {
		err = fmt.Errorf("failed to validate claims: %w", err)
		slog.DebugContext(ctx, err.Error())
	}

	if claims != nil && claims.SpotifyToken != nil {
		http.Redirect(w, r, "/app/dashboard", http.StatusTemporaryRedirect)
		return
	}

	loginPage := LoginPage(LoginPageProps{
		authUrl: h.spotifyAuth.AuthURL(a.Config().StateCode),
	})
	loginPage.Render(r.Context(), w)
}

func (h *HttpHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		httpx.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	a.DeleteClaims(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *HttpHandler) AuthorizeSpotify(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		httpx.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	client, err := h.spotifyAuth.GetClientWithCallback(ctx, a.Config().StateCode, r)
	if err != nil {
		err = fmt.Errorf("failed to get spotify client: %w", err)
		httpx.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	token, err := client.Token()
	if err != nil {
		err = fmt.Errorf("failed to get spotify token: %w", err)
		httpx.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	claims := a.Claims()
	if claims == nil {
		claims = app.NewClaims()
	}
	claims.SpotifyToken = token
	err = a.SetClaims(w, claims)
	if err != nil {
		err = fmt.Errorf("failed to update JWT with Spotify token: %w", err)
		httpx.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
	ctx = ctx.WithApp(a)

	http.Redirect(w, r.WithContext(ctx), "/", http.StatusSeeOther)
}
