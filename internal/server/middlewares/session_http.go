package middlewares

import (
	"context"
	"goph_keeper/internal/server/utils"
	"goph_keeper/internal/shared/models"
	"log/slog"
	"net/http"
	"strings"
)

func JWTSession(secretKey string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			if token == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")
			token = strings.TrimSpace(token)

			userName, err := utils.ParseToken(token, secretKey)
			if err != nil {
				slog.Error("parsing token", "err", err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), models.UserContextKey, userName)

			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}
