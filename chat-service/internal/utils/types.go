package utils

import (
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	GetChats(user uuid.UUID) ([]Chat, error)
	GetChat(id uuid.UUID) (*Chat, error)
	GetChatMessages(chatId uuid.UUID, limit, offset int) ([]Message, error)
	MemberOfChat(userId uuid.UUID, chatId uuid.UUID) error
	GetMessage(messageId uuid.UUID) (Message, error)
	SaveMessage(message Message) error
	CreateOrUpdateChat(chat Chat) error
	UpdateChatActivity(chat uuid.UUID) error
	DeleteChat(chat uuid.UUID) error
	UpdateMessage(message Message) error
	DeleteMessage(message uuid.UUID) error
}

type AuthService interface {
	VerifyToken(token string) (*User, error)
	Exists(ids ...uuid.UUID) (bool, error)
}

type AiService interface {
	AskAI(prompt string, response func(response GenerateResponse)) error
	GuessWords(topic string) ([]string, error)
}

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

type Summary struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	Summary   string    `json:"summary" bson:"summary"`
	ChatID    uuid.UUID `json:"chat_id" bson:"chat_id"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type AiAnswer struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	Question  string    `json:"question" bson:"question"`
	Content   string    `json:"content" bson:"content"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type GenerateResponse struct {

	// Response is the textual response itself.
	Response string `json:"response"`

	// Done specifies if the response is complete.
	Done bool `json:"done"`

	// DoneReason is the reason the model stopped generating text.
	DoneReason string `json:"done_reason,omitempty"`
}
