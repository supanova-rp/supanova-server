package domain

import (
	"context"
)

//go:generate moq -out ../handlers/mocks/user_mock.go -pkg mocks . UserRepository

type UserRepository interface {
	GetUser(context.Context, string) (*User, error)
	GetUsersAndAssignedCourses(context.Context) ([]UserWithAssignedCourses, error)
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AssignedCourseTitle struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type UserWithAssignedCourses struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Email   string                `json:"email"`
	Courses []AssignedCourseTitle `json:"courses"`
}
