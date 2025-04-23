package ws

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	Content   string    `json:"content"`
	SenderID  uuid.UUID `json:"sender"`
	ChatID    uuid.UUID `json:"chat_id"`
	Timestamp time.Time `json:"timestamp"`
}

type SendMessage struct {
	Content string    `json:"content"`
	ChatID  uuid.UUID `json:"chat_id"`
	ID      uuid.UUID `json:"id"`
	Deleted bool      `json:"Deleted"`
}
