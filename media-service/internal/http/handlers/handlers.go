package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/media-service/internal/media"
	"team6-managing.mni.thm.de/Commz/media-service/internal/utils"
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

func NewHandlers(mda *media.MediaService) *Handlers {
	return &Handlers{
		handlers: []Handler{
			&MediaHandler{mda: mda},
		},
	}
}

func HandleFunc(router *mux.Router, path string, handler func(http.ResponseWriter,
	*http.Request), methods ...string) *mux.Route {
	logger.Debug().Str("path", path).Strs("methods", methods).Msg("Registering route")
	if len(methods) == 0 {
		return router.HandleFunc(path, handler)
	}
	return router.HandleFunc(path, handler).Methods(methods...)
}

func (h *Handlers) RegisterRoutes(router *mux.Router) {
	for _, handler := range h.handlers {
		handler.RegisterRoutes(router)
	}
}
