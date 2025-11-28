package testhelpers

import (
	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/domain"
)

var Course = &domain.Course{
	ID:                uuid.New(),
	Title:             "Test Course",
	Description:       "Test Description",
	CompletionTitle:   "Completion Title",
	CompletionMessage: "Completion Message",
	Sections:          []domain.CourseSection{},
	Materials:         []domain.CourseMaterial{},
}
