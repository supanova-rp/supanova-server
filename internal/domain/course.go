package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/course_mock.go -pkg mocks . CourseRepository

type CourseRepository interface {
	GetCourse(context.Context, pgtype.UUID) (*Course, error)
	AddCourse(context.Context, sqlc.AddCourseParams) (*Course, error)
}

type Course struct {
	ID                uuid.UUID        `json:"id"`
	Title             string           `json:"title"`
	Description       string           `json:"description"`
	CompletionTitle   string           `json:"completionTitle"`
	CompletionMessage string           `json:"completionMessage"`
	Sections          []CourseSection  `json:"sections"`
	Materials         []CourseMaterial `json:"materials"`
}

type CourseMaterial struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Position   int       `json:"position"`
	StorageKey uuid.UUID `json:"storageKey"`
}

type CourseSection interface {
	GetID() uuid.UUID
	GetTitle() string
	GetPosition() int
	GetType() SectionType
}

type SectionType string

const SectionTypeVideo SectionType = "video"
const SectionTypeQuiz SectionType = "quiz"

type VideoSection struct {
	ID         uuid.UUID   `json:"id"`
	Title      string      `json:"title"`
	Position   int         `json:"position"`
	StorageKey uuid.UUID   `json:"storageKey"`
	Type       SectionType `json:"type"`
}

// Implements CourseSection interface
func (v *VideoSection) GetID() uuid.UUID     { return v.ID }
func (v *VideoSection) GetTitle() string     { return v.Title }
func (v *VideoSection) GetPosition() int     { return v.Position }
func (v *VideoSection) GetType() SectionType { return v.Type }

type QuizSection struct {
	ID        uuid.UUID      `json:"id"`
	Title     string         `json:"title"`
	Position  int            `json:"position"`
	Type      SectionType    `json:"type"`
	Questions []QuizQuestion `json:"questions"`
}

// Implements CourseSection interface
func (q *QuizSection) GetID() uuid.UUID     { return q.ID }
func (q *QuizSection) GetTitle() string     { return q.Title }
func (q *QuizSection) GetPosition() int     { return q.Position }
func (q *QuizSection) GetType() SectionType { return q.Type }

type QuizQuestion struct {
	ID            uuid.UUID    `json:"id"`
	Question      string       `json:"question"`
	Position      int          `json:"position"`
	IsMultiAnswer bool         `json:"isMultiAnswer"`
	Answers       []QuizAnswer `json:"answers"`
}

type QuizAnswer struct {
	ID              uuid.UUID `json:"id"`
	Answer          string    `json:"answer"`
	Position        int       `json:"position"`
	IsCorrectAnswer bool      `json:"isCorrectAnswer"`
}
