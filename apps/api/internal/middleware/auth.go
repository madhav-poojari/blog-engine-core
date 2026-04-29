package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/typewriter-blog/api/internal/service"
)

type contextKey string

const AdminIDKey contextKey = "adminID"

// Auth returns middleware that validates JWT tokens on protected routes.
// If the token is valid, the admin ID is stored in the request context.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AdminIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth extracts the admin ID from a valid JWT if present,
// but does NOT reject requests without a token.
func OptionalAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header != "" && strings.HasPrefix(header, "Bearer ") {
				tokenStr := strings.TrimPrefix(header, "Bearer ")
				claims, err := authService.ValidateToken(tokenStr)
				if err == nil {
					ctx := context.WithValue(r.Context(), AdminIDKey, claims.Subject)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IsAdmin checks whether the current request has a valid admin context.
func IsAdmin(ctx context.Context) bool {
	_, ok := ctx.Value(AdminIDKey).(string)
	return ok
}
