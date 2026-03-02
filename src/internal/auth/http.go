package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/core/httpx"
	"shmoopicks/src/internal/feed"
	"shmoopicks/src/internal/spotify"
	"shmoopicks/src/internal/user"
)

type HttpHandler struct {
	spotifyAuth *spotify.AuthService
	userService *user.Service
	feedService *feed.Service
}

func NewHttpHandler(spotifyAuth *spotify.AuthService, userService *user.Service, feedService *feed.Service) *HttpHandler {
	return &HttpHandler{
		spotifyAuth: spotifyAuth,
		userService: userService,
		feedService: feedService,
	}
}

func (h *HttpHandler) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	claims, err := app.ValidateClaimsFromRequest(r, a.Config().JwtSecret)
	if err != nil {
		err = fmt.Errorf("failed to validate claims: %w", err)
		slog.DebugContext(ctx, err.Error())
	}

	if claims != nil && claims.UserID != nil {
		user, err := h.userService.GetUserById(ctx, *claims.UserID)
		if err != nil {
			err = fmt.Errorf("failed to get user: %w", err)
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}

		if user.SpotifyRefreshToken(a.Config().SpotifyTokenSecret) != nil {
			http.Redirect(w, r, "/app/library/dashboard", http.StatusTemporaryRedirect)
			return
		}
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
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
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
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	client, err := h.spotifyAuth.GetClientFromCallback(ctx, a.Config().StateCode, r)
	if err != nil {
		err = fmt.Errorf("failed to get spotify client: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	token, err := client.Token()
	if err != nil {
		err = fmt.Errorf("failed to get spotify token: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	spotifyUser, err := client.CurrentUser(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get spotify user: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
	user, err := h.userService.UpsertSpotifyUser(ctx, spotifyUser.ID, token.RefreshToken)
	if err != nil {
		err = fmt.Errorf("failed to upsert spotify user: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	_, err = h.feedService.UpsertFeed(ctx, user.ID, models.FeedKindSpotify)
	if err != nil {
		err = fmt.Errorf("failed to upsert feed: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	claims := a.Claims()
	if claims == nil {
		claims = app.NewClaims()
	}
	claims.UserID = &user.ID
	err = a.SetClaims(w, claims)
	if err != nil {
		err = fmt.Errorf("failed to update JWT with user ID: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
	ctx = ctx.WithApp(a)

	http.Redirect(w, r.WithContext(ctx), "/", http.StatusSeeOther)
}
