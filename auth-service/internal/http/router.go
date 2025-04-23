package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/auth"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/http/handlers"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"
)

var (
	logger = utils.GetLogger("http")
)

type Router struct {
	Router  *mux.Router
	handler *handlers.Handlers
}

func New(auth *auth.AuthService) *Router {
	logger.Info().Msg("Registering routes")

	router := mux.NewRouter()
	handlers := handlers.NewHandlers(auth)

	start := time.Now()
	handlers.RegisterRoutes(router)
	logger.Info().Msgf("Routes registered in %s", time.Since(start))

	// logging middleware
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().Msgf("Request: %s %s", r.Method, r.URL.Path)
			h.ServeHTTP(w, r)
		})
	})

	return &Router{
		Router:  router,
		handler: handlers,
	}
}
