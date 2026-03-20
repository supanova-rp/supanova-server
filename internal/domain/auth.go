package domain

import "context"

type AuthRepository interface {
	RegisterUser(context.Context, RegisterParams) (*User, error)
}

type RegisterParams struct {
	ID    string
	Name  string
	Email string
}
