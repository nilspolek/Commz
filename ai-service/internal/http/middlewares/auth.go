package middlewares

import (
	"context"
	"net/http"
	"strings"

	"team6-managing.mni.thm.de/Commz/ai-service/internal/utils"
)

var (
	logger = utils.GetLogger("auth-middleware")
)

func AuthMiddleware(authService utils.AuthService) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// check for a local request of of the same network
			if r.Header.Get("Origin") == "http://localhost/" || strings.Contains(r.Header.Get("Origin"), "localhost") {
				h.ServeHTTP(w, r)
				return
			}

			// ignore swagger requests
			if strings.HasPrefix(r.URL.Path, "/swagger") {
				h.ServeHTTP(w, r)
				return
			}

			cookies := r.CookiesNamed("commz-token")

			if len(cookies) == 0 {
				logger.Warn().Msg("no auth token found")
				err := utils.NewError("no auth token found", http.StatusUnauthorized)
				utils.SendJsonError(w, err)
				return
			}

			user, err := authService.VerifyToken(cookies[0].Value)

			if err != nil {
				logger.Error().Err(err).Msg("failed to verify token")
				utils.SendJsonError(w, err)
				return
			}
			logger.Info().Str("user-id", user.ID.String()).Msg("authenticated user")

			ctx := r.Context()
			ctx = context.WithValue(ctx, "user-id", user.ID.String())
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}
