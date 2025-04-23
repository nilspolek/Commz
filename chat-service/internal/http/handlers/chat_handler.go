package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nilspolek/DevOps/Chat/internal/chat"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
)

type ChatHandler struct {
	chat *chat.ChatService
}

func (c *ChatHandler) RegisterRoutes(router *mux.Router) {

	// Chat endpoints
	HandleFunc(router, "/", c.getChats, "GET")
	HandleFunc(router, "/", c.createChat, "POST")

	HandleFunc(router, "/version", c.getVersion, "GET")

	HandleFunc(router, "/{chatId}", c.getChat, "GET")
	HandleFunc(router, "/{chatId}", c.updateChat, "PUT")
	HandleFunc(router, "/{chatId}", c.deleteChat, "DELETE")

	HandleFunc(router, "/{chatId}/messages", c.getChatMessages, "GET")
	HandleFunc(router, "/{chatId}/messages", c.sendChatMessage, "POST")
	HandleFunc(router, "/messages/{messageId}", c.updateChatMessage, "PUT")
	HandleFunc(router, "/messages/{messageId}", c.deleteChatMessage, "DELETE")
	HandleFunc(router, "/messages/{messageId}/read", c.readMessage, "GET")

	HandleFunc(router, "/direct-chat", c.createDirectChat, "POST")
}

// @Summary Get the service Version
// @Success 200 {object} string "version"
// @Router /version [get]
func (c *ChatHandler) getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(utils.VERSION))
}

// @Summary Sets the status of a message to read
// @Description returns the updated message
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param messageId path string true "Message ID"
// @Success 200 {object} SendMessageRequest "Message Updated"
// @Failure 400 {object} utils.ServiceError "Invalid request body or Chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Chat not found or message not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /messages/{messageId}/read [get]
// @Security ApiKeyAuth
func (c *ChatHandler) readMessage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	messageId := mux.Vars(r)["messageId"]
	messageUUID, err := uuid.Parse(messageId)
	if err != nil {
		c.error(w, "Invalid message id", http.StatusBadRequest)
		return
	}

	updated, err := c.chat.ReadMessage(userId, messageUUID)
	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, updated)
}

// @Summary Update a message by id
// @Description returns the updated message
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param messageId path string true "Message ID"
// @Success 200 {object} SendMessageRequest "Message Updated"
// @Failure 400 {object} utils.ServiceError "Invalid request body or Chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /messages/{messageId} [put]
// @Security ApiKeyAuth
func (c *ChatHandler) updateChatMessage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	messageId := mux.Vars(r)["messageId"]
	messageUUID, err := uuid.Parse(messageId)
	if err != nil {
		c.error(w, "Invalid message id", http.StatusBadRequest)
		return
	}

	// get message from request
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, "Invalid message", http.StatusBadRequest)
		return
	}

	// parse message
	var message SendMessageRequest
	err = json.Unmarshal(b, &message)
	if err != nil {
		c.error(w, "Invalid message", http.StatusBadRequest)
		return
	}

	updated, err := c.chat.UpdateMessage(userId, utils.Message{
		Content: message.Message,
		ID:      messageUUID,
		Media:   message.Media,
	})
	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, updated)
}

// @Summary Delete a message by id
// @Description returns the deleted message
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param messageId path string true "Message ID"
// @Success 200 {object} utils.Message "Message Deleted"
// @Failure 400 {object} utils.ServiceError "Invalid request body or Chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /messages/{messageId} [delete]
// @Security ApiKeyAuth
func (c *ChatHandler) deleteChatMessage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	messageId := mux.Vars(r)["messageId"]
	messageUUID, err := uuid.Parse(messageId)

	if err != nil {
		c.error(w, "Invalid message id", http.StatusBadRequest)
		return
	}

	message, err := c.chat.DeleteMessage(userId, messageUUID)
	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, message)
}

// @Summary Gets a specific chat by ID
// @Description Returns the chat with that ID
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param chatId path string true "Chat ID"
// @Success 200 {object} utils.Chat "Chat Found"
// @Failure 400 {object} utils.ServiceError "Invalid request body or Chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /{chatId} [get]
// @Security ApiKeyAuth
func (c *ChatHandler) getChat(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	chatId := mux.Vars(r)["chatId"]
	chatIdUUID, err := uuid.Parse(chatId)

	if err != nil {
		c.error(w, "Invalid chat id", http.StatusBadRequest)
		return
	}

	// create chat
	chat, err := c.chat.GetChat(chatIdUUID)

	if c.handleErrors(err, w) {
		return
	}

	if !slices.ContainsFunc(chat.Members, func(id uuid.UUID) bool { return id.String() == userId.String() }) {
		c.error(w, "Unauthorized: User is not a member of this chat", http.StatusForbidden)
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, chat)
}

// @Summary Delete a chat
// @Description Deletes an exsisting chat, the user has to be part of that chat.
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param chatId path string true "Chat ID"
// @Success 200 {object} bool "Chat deleted"
// @Failure 400 {object} utils.ServiceError "Invalid request body or chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /{chatId} [delete]
// @Security ApiKeyAuth
func (c *ChatHandler) deleteChat(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	chatId := mux.Vars(r)["chatId"]
	chatIdUUID, err := uuid.Parse(chatId)
	if err != nil {
		c.error(w, "Invalid chat id", http.StatusBadRequest)
		return
	}

	// delete chat
	err = c.chat.DeleteChat(userId, chatIdUUID)
	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, true)
}

// @Summary Updates a chat between users
// @Description Updates achat between the authenticated user other users
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param chatId path string true "Chat ID"
// @Param request body CreateChatRequest true "Direct chat creation request"
// @Success 200 {object} utils.Chat "Chat creation successful"
// @Failure 400 {object} utils.ServiceError "Invalid request body or chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /{chatId} [put]
// @Security ApiKeyAuth
func (c *ChatHandler) updateChat(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat from request
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, "Invalid chat", http.StatusBadRequest)
		return
	}
	// parse chat
	var chat CreateChatRequest
	err = json.Unmarshal(b, &chat)
	if err != nil {
		logger.Error().Err(err).Msg("Error parsing chat")
		c.error(w, "Invalid create chat request", http.StatusBadRequest)
		return
	}

	// get chat id from request
	chatId := mux.Vars(r)["chatId"]
	chatIdUUID, err := uuid.Parse(chatId)
	if err != nil {
		c.error(w, "Invalid chat id", http.StatusBadRequest)
		return
	}

	// create chat
	chatResponse, err := c.chat.UpdateChat(userId, chatIdUUID, chat.Name, chat.Members)

	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, chatResponse)
}

// @Summary Create a direct chat between two users
// @Description Creates a new direct chat between the authenticated user and another user
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param userid path string true "Receiver user ID"
// @Param request body StartDirectMessageRequest true "Direct chat creation request"
// @Success 200 {object} utils.Chat "Chat creation successful"
// @Failure 400 {object} utils.ServiceError "Invalid request body or user ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "User not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /direct-chat [post]
// @Security ApiKeyAuth
func (c *ChatHandler) createDirectChat(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat from request
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, "Invalid chat", http.StatusBadRequest)
		return
	}

	// parse chat
	var chat StartDirectMessageRequest
	err = json.Unmarshal(b, &chat)
	if err != nil {
		c.error(w, "Invalid create chat request", http.StatusBadRequest)
		return
	}

	// create chat
	chatResponse, err := c.chat.CreateDirectChat(userId, chat.Receiver, chat.Message)

	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, chatResponse)
}

// @Summary Create a group chat
// @Description Creates a new group chat with multiple members
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param request body CreateChatRequest true "Group chat creation request"
// @Success 200 {object} utils.Chat "Chat creation successful"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "One or more users not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router / [post]
// @Security ApiKeyAuth
func (c *ChatHandler) createChat(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat from request
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, "Invalid chat", http.StatusBadRequest)
		return
	}
	// parse chat
	var chat CreateChatRequest
	err = json.Unmarshal(b, &chat)
	if err != nil {
		logger.Error().Err(err).Msg("Error parsing chat")
		c.error(w, "Invalid create chat request", http.StatusBadRequest)
		return
	}

	// create chat
	chatResponse, err := c.chat.CreateChat(userId, chat.Name, chat.Members, chat.Message)

	if c.handleErrors(err, w) {
		return
	}

	// send chat as response
	utils.SendJsonResponse(w, chatResponse)
}

// @Summary Get user's chats
// @Description Retrieves all chats (both direct and group) that the authenticated user is a member of
// @Tags chat
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Success 200 {array} utils.Chat "List of chats"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router / [get]
// @Security ApiKeyAuth
func (c *ChatHandler) getChats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	chats, err := c.chat.GetChats(userId)

	if c.handleErrors(err, w) {
		return
	}

	// send chats as response
	utils.SendJsonResponse(w, chats)
}

// @Summary Get chat messages
// @Description Retrieves all messages from a specific chat
// @Tags chat
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param chatId path string true "Chat ID" format(uuid)
// @Param limit query int false "Number of messages to return" default(50)
// @Param offset query int false "Number of messages to skip" default(0)
// @Success 200 {array} utils.Message "List of chat messages"
// @Failure 400 {object} utils.ServiceError "Invalid chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 403 {object} utils.ServiceError "User not member of chat"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /{chatId}/messages [get]
// @Security ApiKeyAuth
func (c *ChatHandler) getChatMessages(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	chatId := mux.Vars(r)["chatId"]
	chatIdUUID, err := uuid.Parse(chatId)

	if err != nil {
		c.error(w, "Invalid chat id", http.StatusBadRequest)
		return
	}

	// Parse limit and offset from query parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	messages, err := c.chat.GetMessages(userId, chatIdUUID, limit, offset)

	if c.handleErrors(err, w) {
		return
	}

	// send messages as response
	utils.SendJsonResponse(w, messages)
}

// @Summary Send chat message
// @Description Sends a new message to a specific chat
// @Tags chat
// @Accept json
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Param chatId path string true "Chat ID" format(uuid)
// @Param request body SendMessageRequest true "Message content"
// @Success 200 {object} utils.Message "Message sent successfully"
// @Failure 400 {object} utils.ServiceError "Invalid request body or chat ID"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 403 {object} utils.ServiceError "User not member of chat"
// @Failure 404 {object} utils.ServiceError "Chat not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /{chatId}/messages [post]
// @Security ApiKeyAuth
func (c *ChatHandler) sendChatMessage(w http.ResponseWriter, r *http.Request) {
	// get user id from context
	user := r.Context().Value("user-id")
	userId := uuid.MustParse(user.(string))

	// get chat id from request
	chatId := mux.Vars(r)["chatId"]
	chatIdUUID, err := uuid.Parse(chatId)

	if err != nil {
		c.error(w, "Invalid chat id", http.StatusBadRequest)
		return
	}

	// get message from request
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, "Invalid message", http.StatusBadRequest)
		return
	}

	// parse message
	var message SendMessageRequest
	err = json.Unmarshal(b, &message)
	if err != nil || message.Message == "" && len(message.Media) == 0 {
		c.error(w, "Invalid message", http.StatusBadRequest)
		return
	}

	if message.Command != "" {
		// handle command
		messageResult, err := c.chat.Command(userId, chatIdUUID, message.Message, message.Command)
		if c.handleErrors(err, w) {
			return
		}

		// send success response
		utils.SendJsonResponse(w, messageResult)
		return
	}

	messageResult, err := c.chat.SendMessage(userId, chatIdUUID, message.Message, message.Media, message.ReplyTo)
	if c.handleErrors(err, w) {
		return
	}

	// send success response
	utils.SendJsonResponse(w, messageResult)
}

func (c *ChatHandler) handleErrors(err error, w http.ResponseWriter) bool {
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

func (c *ChatHandler) error(w http.ResponseWriter, error string, code int) {
	err := utils.NewError(error, code)
	c.handleErrors(err, w)
}
