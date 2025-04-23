package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/ollama/ollama/api"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/ai"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/auth"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/utils"
)

var CodesToLog = map[int]bool{
	http.StatusBadRequest:          true,
	http.StatusUnauthorized:        true,
	http.StatusNotFound:            true,
	http.StatusInternalServerError: true,
}

type AiHandler struct {
	ai   *ai.AiService
	auth *auth.AuthService
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (c *AiHandler) RegisterRoutes(router *mux.Router) {
	HandleFunc(router, "/summarization", c.summarizeChat)
	HandleFunc(router, "/fix", c.correctText)
	HandleFunc(router, "/rewrite", c.rewriteText)
	HandleFunc(router, "/ask", c.askAi)
	HandleFunc(router, "/guess", c.guessWord).Methods("POST")

	HandleFunc(router, "/version", c.getVersion).Methods("GET")

}

// @Summary Get the service Version
// @Success 200 {object} string "version"
// @Router /version [get]
func (c *AiHandler) getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(utils.VERSION))
}

// @Summary Gives a list of words for that specific topic
// @Description Gives a list of words for that specific topic
// @Tags ai
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param request body TextManipulationRequest true "the topic for the words to generate"
// @Success 200 {object} []string "generated words"
// @Failure 400 {object} utils.ServiceError "Invalid request body or user ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /guess [post]
func (c *AiHandler) guessWord(w http.ResponseWriter, r *http.Request) {
	var request TextManipulationRequest
	json.NewDecoder(r.Body).Decode(&request)

	resp, err := c.ai.GenerateGuessWords(request.Text)
	if c.handleErrors(err, w) {
		return
	}
	utils.SendJsonResponse(w, resp)
}

func (c *AiHandler) summarizeChat(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		c.error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	var chat utils.Chat
	if err = conn.ReadJSON(&chat); err != nil {
		c.error(w, "Failed to read message", http.StatusInternalServerError)
	}

	err = c.ai.SummarizeChat(chat, func(resp api.GenerateResponse) error {
		return conn.WriteJSON(resp)
	})
	c.handleErrors(err, w)
}

func (c *AiHandler) askAi(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		c.error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	var request AskAiRequest
	if err = conn.ReadJSON(&request); err != nil {
		c.error(w, "Failed to read message", http.StatusInternalServerError)
		return
	}

	err = c.ai.AskAI(request.Prompt, func(resp api.GenerateResponse) error {
		return conn.WriteJSON(resp)
	})
	c.handleErrors(err, w)
}

func (c *AiHandler) correctText(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	if err != nil {
		c.error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	var request TextManipulationRequest
	if err = conn.ReadJSON(&request); err != nil {
		c.error(w, "Failed to read message", http.StatusInternalServerError)
		return
	}

	err = c.ai.CorrectText(request.Text, func(resp api.GenerateResponse) error {
		return conn.WriteJSON(resp)
	})
	c.handleErrors(err, w)
}

func (c *AiHandler) rewriteText(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	if err != nil {
		c.error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	var request TextManipulationRequest
	if err = conn.ReadJSON(&request); err != nil {
		c.error(w, "Failed to read message", http.StatusInternalServerError)
		return
	}

	err = c.ai.RewriteText(request.Text, func(resp api.GenerateResponse) error {
		return conn.WriteJSON(resp)
	})
	c.handleErrors(err, w)
}

func (c *AiHandler) handleErrors(err error, w http.ResponseWriter) bool {
	// check if error is a custom error
	customError, ok := err.(*utils.ServiceError)
	if ok {
		// marshal error into json and send status from error as statuscode
		w.WriteHeader(customError.StatusCode)
		w.Write(customError.Bytes())
		return true
	}

	if err != nil {
		customError = &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Err:        "Internal server error",
			Message:    err.Error(),
		}
		utils.SendJsonError(w, customError)
		return true
	}
	return false
}

func (c *AiHandler) error(w http.ResponseWriter, error string, code int) {
	err := utils.NewError(error, code)
	if CodesToLog[code] {
		logger.Error().Err(err)
	}
	c.handleErrors(err, w)
}
