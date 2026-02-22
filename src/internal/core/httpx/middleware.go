package httpx

import (
	"fmt"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"time"
)

type Middleware func(HandlerFunc) HandlerFunc

func ApplyMiddleware(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func JwtMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {

		claims, err := appctx.ValidateClaimsFromRequest(r, ctx.Config().JwtSecret)
		if err != nil || claims.SpotifyToken == nil {
			ctx.DeleteClaims(w)
			HandleErrorResponse(ctx, w, http.StatusUnauthorized, fmt.Errorf("Invalid or expired token: %s", err.Error()))
			return
		}

		// Add claims to request context
		err = ctx.SetClaims(w, *claims)
		if err != nil {
			HandleErrorResponse(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to set JWT: %w", err))
			return
		}

		// Call next handler
		next(ctx, w, r)
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
	return func(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
		ww := &RequestLoggingMiddlewareResponseWriter{ResponseWriter: w, statusCode: 200, startTime: time.Now()}
		next(ctx, ww, r)
		slog.InfoContext(ctx, "Request", "status", ww.statusCode, "method", r.Method, "path", r.URL.Path, "url", r.URL.String(), "duration", ww.Duration())
	}
}
