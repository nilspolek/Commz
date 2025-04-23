package utils

import (
	"time"

	"github.com/google/uuid"
)

const VERSION = "1.2.0"

type Message struct {
	ID        uuid.UUID   `json:"id" bson:"_id"`
	Content   string      `json:"content" bson:"content"`
	Command   string      `json:"command" bson:"command"`
	SenderID  uuid.UUID   `json:"sender" bson:"sender"`
	ChatID    uuid.UUID   `json:"chat_id" bson:"chat_id"`
	UpdatedAt time.Time   `json:"updatedAt" bson:"updatedAt"`
	Timestamp time.Time   `json:"timestamp" bson:"timestamp"`
	Media     []uuid.UUID `json:"media" bson:"media"`
	Read      bool        `json:"read" bson:"read"`
	ReplyTo   *uuid.UUID  `json:"reply_to" bson:"reply_to"`
	Deleted   bool        `json:"deleted" bson:"deleted"`
}

type Chat struct {
	ID         uuid.UUID   `json:"id" bson:"_id"`
	Name       string      `json:"name" bson:"name"`
	Members    []uuid.UUID `json:"members" bson:"members"`
	Messages   []Message   `json:"messages" bson:"-"`
	CreatorID  uuid.UUID   `json:"creator_id" bson:"creator_id"`
	CreatedAt  time.Time   `json:"created_at" bson:"created_at"`
	LastActive time.Time   `json:"last_active" bson:"last_active"`
}

type User struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	Password  string    `json:"password" bson:"password"`
	Email     string    `json:"email" bson:"email"`
	FirstName string    `json:"first_name" bson:"first_name"`
	LastName  string    `json:"last_name" bson:"last_name"`
}
