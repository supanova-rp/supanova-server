package domain

import "github.com/google/uuid"

type Course struct {
	ID          uuid.UUID
	Title       string
	Description string
}
