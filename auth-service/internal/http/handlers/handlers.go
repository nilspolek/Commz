package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/auth"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"
)

var (
	logger = utils.GetLogger("routes")
)

type Handler interface {
	RegisterRoutes(router *mux.Router)
}

type Handlers struct {
	handlers []Handler
}

func NewHandlers(auth *auth.AuthService) *Handlers {
	return &Handlers{
		handlers: []Handler{
			&AuthHandler{auth: auth},
		},
	}
}

func HandleFunc(router *mux.Router, path string, handler func(http.ResponseWriter,
	*http.Request), methods ...string) *mux.Route {
	logger.Debug().Str("path", path).Strs("methods", methods).Msg("Registering route")
	return router.HandleFunc(path, handler).Methods(methods...)
}

func (h *Handlers) RegisterRoutes(router *mux.Router) {
	for _, handler := range h.handlers {
		handler.RegisterRoutes(router)
	}
}
