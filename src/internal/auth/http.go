package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"shmoopicks/src/internal/core/apphttp"
	"shmoopicks/src/internal/core/db"
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

func (h *HttpHandler) GetLoginPage(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
	claims, err := appctx.ValidateClaimsFromRequest(r, ctx.Config().JwtSecret)
	if err != nil {
		slog.Error(err.Error())
	}

	if claims != nil && claims.SpotifyToken != nil {
		http.Redirect(w, r, "/app/dashboard", http.StatusTemporaryRedirect)
		return
	}

	loginPage := LoginPage(LoginPageProps{
		authUrl: h.spotifyAuth.AuthURL(ctx.Config().StateCode),
	})
	loginPage.Render(r.Context(), w)
}

func (h *HttpHandler) Logout(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
	ctx.DeleteJwt(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *HttpHandler) AuthorizeSpotify(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
	client, err := h.spotifyAuth.GetClientWithCallback(ctx, ctx.Config().StateCode, r)
	if err != nil {
		apphttp.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	token, err := client.Token()
	if err != nil {
		apphttp.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if !ctx.HasJwt() {
		err = ctx.SetJwt(w, *appctx.NewClaims())
		if err != nil {
			err = fmt.Errorf("failed to set JWT: %w", err)
			apphttp.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}
	}

	err = ctx.UpdateJwt(w, func(jwt appctx.Claims) appctx.Claims {
		jwt.SpotifyToken = token
		return jwt
	})
	if err != nil {
		err = fmt.Errorf("failed to update JWT with Spotify token: %w", err)
		apphttp.HandleErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
	slog.Info("JWT updated with Spotify token", "token", token)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
