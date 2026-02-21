package apphttp

import (
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"strings"
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
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := appctx.ValidateJWT(tokenString, ctx.Config().JwtSecret)
		if err != nil {
			http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx.SetJwt(*claims)

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
	slog.Info("Initializing logging middleware")
	return func(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
		ww := &RequestLoggingMiddlewareResponseWriter{ResponseWriter: w, statusCode: 200, startTime: time.Now()}
		next(ctx, ww, r)
		slog.InfoContext(ctx, "Request", "status", ww.statusCode, "method", r.Method, "path", r.URL.Path, "url", r.URL.String(), "duration", ww.Duration())
	}
}
