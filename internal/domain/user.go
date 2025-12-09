package domain

import (
	"context"
)

//go:generate moq -out ../domain/mocks/user_mock.go -pkg mocks . UserRepository

type UserRepository interface {
	GetUser(context.Context, string) (*User, error)
}

type User struct {
	ID    string
	Name  string `json:"name"`
	Email string `json:"email"`
}
