package apphttp

import (
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"strings"
)

type Middleware func(AppHandlerFunc) AppHandlerFunc

func ApplyMiddleware(handler AppHandlerFunc, middlewares ...Middleware) AppHandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func JwtMiddleware(next AppHandlerFunc) AppHandlerFunc {
	return func(ctx appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
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
