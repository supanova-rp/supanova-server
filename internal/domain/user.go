package domain

import (
	"context"
)

type UserRepository interface {
	GetUser(context.Context, string) (*User, error)
}

type User struct {
	ID    string
	Name  string `json:"name"`
	Email string `json:"email"`
}
