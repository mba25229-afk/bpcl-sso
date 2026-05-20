package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bpcl/portal-api/internal/service"
)

// TokenValidator is implemented by *service.AuthService.
type TokenValidator interface {
	ValidateToken(tokenStr string) (*service.Claims, error)
}

func Auth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeUnauth(w)
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validator.ValidateToken(tokenStr)
			if err != nil {
				writeUnauth(w)
				return
			}
			ctx := context.WithValue(r.Context(), service.ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "code": "UNAUTHORIZED"})
}
