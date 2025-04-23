package http

import (
	"time"

	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/ai"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/http/handlers"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/http/middlewares"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/utils"
)

var (
	logger = utils.GetLogger("http")
)

type Router struct {
	Router  *mux.Router
	handler *handlers.Handlers
}

func New(ai *ai.AiService, auth utils.AuthService) *Router {
	logger.Info().Msg("Registering routes")

	router := mux.NewRouter()
	handlers := handlers.NewHandlers(ai)
	start := time.Now()
	handlers.RegisterRoutes(router)
	logger.Info().Msgf("Routes registered in %s", time.Since(start))

	router.Use(mux.CORSMethodMiddleware(router))

	// logging middleware
	router.Use(middlewares.LoggingMiddleware(logger))
	router.Use(middlewares.AuthMiddleware(auth))

	return &Router{
		Router:  router,
		handler: handlers,
	}
}
