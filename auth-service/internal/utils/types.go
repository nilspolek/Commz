package utils

import (
	"github.com/google/uuid"
)

type Storage interface {
	UpdateOrCreateUser(user User) error
	GetUserByEmail(email string) (User, error)
	Exists(email string) bool
	GetUserByID(id uuid.UUID) (User, error)
	GetUsers() ([]User, error)
}

type User struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	Password  string    `json:"-" bson:"password"`
	Email     string    `json:"email" bson:"email"`
	FirstName string    `json:"first_name" bson:"first_name"`
	LastName  string    `json:"last_name" bson:"last_name"`
	Picture   string    `json:"picture" bson:"picture"`
}
