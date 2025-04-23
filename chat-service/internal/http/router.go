package http

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/nilspolek/DevOps/Chat/internal/chat"
	"github.com/nilspolek/DevOps/Chat/internal/http/handlers"
	"github.com/nilspolek/DevOps/Chat/internal/http/middlewares"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
)

var (
	logger = utils.GetLogger("http")
)

type Router struct {
	Router  *mux.Router
	handler *handlers.Handlers
}

func New(chat *chat.ChatService, authService utils.AuthService) *Router {
	logger.Info().Msg("Registering routes")

	router := mux.NewRouter()
	handlers := handlers.NewHandlers(chat)

	start := time.Now()
	handlers.RegisterRoutes(router)
	logger.Info().Msgf("Routes registered in %s", time.Since(start))

	router.Use(mux.CORSMethodMiddleware(router))

	// logging middleware
	router.Use(middlewares.LoggingMiddleware(logger))

	// user auth middleware
	router.Use(middlewares.AuthMiddleware(authService))

	return &Router{
		Router:  router,
		handler: handlers,
	}
}
