package domain

import (
	"context"

	"github.com/google/uuid"
)

//go:generate moq -out ../handlers/mocks/progress_mock.go -pkg mocks . ProgressRepository

type ProgressRepository interface {
	GetProgress(context.Context, GetProgressParams) (*Progress, error)
	UpdateProgress(context.Context, UpdateProgressParams) error
	GetAllProgress(context.Context) ([]*FullProgress, error)
	HasCompletedCourse(context.Context, HasCompletedCourseParams) (bool, error)
	SetCourseCompleted(context.Context, SetCourseCompletedParams) error
}

type GetProgressParams struct {
	UserID   string
	CourseID uuid.UUID
}

type UpdateProgressParams struct {
	UserID    string
	CourseID  uuid.UUID
	SectionID uuid.UUID
}

type HasCompletedCourseParams struct {
	UserID   string
	CourseID uuid.UUID
}

type SetCourseCompletedParams struct {
	UserID   string
	CourseID uuid.UUID
}

type Progress struct {
	CompletedSectionIDs []uuid.UUID `json:"completedSectionIds"`
	CompletedIntro      bool        `json:"completedIntro"`
}

type FullProgress struct {
	UserID   string              `json:"userID"`
	UserName string              `json:"name"`
	Email    string              `json:"email"`
	Progress []*FullUserProgress `json:"progress"`
}

type FullUserProgress struct {
	CourseID              uuid.UUID               `json:"courseID"`
	CourseName            string                  `json:"courseName"`
	CompletedIntro        bool                    `json:"completedIntro"`
	CompletedCourse       bool                    `json:"completedCourse"`
	CourseSectionProgress []CourseSectionProgress `json:"courseSectionProgress"`
}

type CourseSectionProgress struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	Type      string    `json:"type"`
	Completed bool      `json:"completed"`
}
