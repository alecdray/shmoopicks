package httpx

import (
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/spotify"
	"shmoopicks/src/internal/user"
	"time"
)

type Middleware func(HandlerFunc) HandlerFunc

func ApplyMiddleware(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func JwtMiddleware(spotifyService *spotify.Service, userService *user.Service) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := contextx.NewContextX(r.Context())
			a, err := ctx.App()
			if err != nil {
				HandleErrorResponse(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to get app: %w", err))
				return
			}

			claims, err := app.ValidateClaimsFromRequest(r, a.Config().JwtSecret)
			if err != nil || (claims.SpotifyToken == nil && claims.UserID == nil) {
				a.DeleteClaims(w)
				HandleErrorResponse(ctx, w, http.StatusUnauthorized, fmt.Errorf("Invalid or expired token: %s", err.Error()))
				return
			}

			if claims.UserID != nil {
				ctx = ctx.WithUserId(*claims.UserID)
			}

			// Add claims to request context
			err = a.SetClaims(w, claims)
			if err != nil {
				HandleErrorResponse(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to set JWT: %w", err))
				return
			}
			ctx = ctx.WithApp(a)

			user, err := userService.GetUserFromCtx(ctx)
			if err != nil {
				HandleErrorResponse(ctx, w, http.StatusUnauthorized, fmt.Errorf("failed to get user: %w", err))
				return
			}
			ctx = ctx.WithUserId(user.ID)

			claims.UserID = &user.ID
			err = a.SetClaims(w, claims)
			if err != nil {
				HandleErrorResponse(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to set JWT: %w", err))
				return
			}
			ctx = ctx.WithApp(a)

			r = r.WithContext(ctx)
			// Call next handler
			next(w, r)
		}
	}
}

type RequestLoggingMiddlewareResponseWriter struct {
	http.ResponseWriter
	statusCode int
	startTime  time.Time
}

func (w *RequestLoggingMiddlewareResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *RequestLoggingMiddlewareResponseWriter) Duration() time.Duration {
	return time.Since(w.startTime)
}

func RequestLoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := &RequestLoggingMiddlewareResponseWriter{ResponseWriter: w, statusCode: 200, startTime: time.Now()}
		next(ww, r)
		slog.InfoContext(r.Context(), "Request", "status", ww.statusCode, "method", r.Method, "path", r.URL.Path, "url", r.URL.String(), "duration", ww.Duration())
	}
}
