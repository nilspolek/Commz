package middlewares

import (
	"net/http"

	"github.com/rs/zerolog"
)

func LoggingMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().Msgf("Request: %s %s", r.Method, r.URL.Path)
			h.ServeHTTP(w, r)
		})
	}
}
