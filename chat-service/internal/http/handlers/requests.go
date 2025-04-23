package handlers

import (
	"github.com/google/uuid"
)

// StartDirectMessageRequest represents the request body for creating a direct chat
type StartDirectMessageRequest struct {
	// The ID of the user to start a chat with
	// required: true
	Receiver uuid.UUID `json:"receiver"`
	// The initial message to send
	// required: true
	Message *string `json:"message"`
}

// CreateChatRequest represents the request body for creating a group chat
type CreateChatRequest struct {
	// Name of the group chat
	// required: true
	Name string `json:"name"`
	// List of user IDs to add to the chat
	// required: true
	// minimum items: 1
	Members []uuid.UUID `json:"members"`
	// The initial message to send
	// required: true
	Message *string `json:"message"`
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	// The message content
	// required: true
	// min length: 1
	Message string      `json:"message"`
	Media   []uuid.UUID `json:"media"`
	Command string      `json:"command"`
	ReplyTo *uuid.UUID  `json:"reply_to"`
}
