package domain

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:generate moq -out ../handlers/mocks/course_mock.go -pkg mocks . CourseRepository

type CourseRepository interface {
	GetCourse(context.Context, pgtype.UUID) (*Course, error)
	GetCoursesOverview(context.Context) ([]CourseOverview, error)
	AddCourse(context.Context, *AddCourseParams) (*Course, error)
	DeleteCourse(context.Context, uuid.UUID) error
}

type AddMaterialParams struct {
	ID         uuid.UUID
	Name       string
	StorageKey uuid.UUID
	Position   int
}

type AddSectionQuestionAnswerParams struct {
	Answer          string
	IsCorrectAnswer bool
	Position        int
}

type AddSectionQuestionParams struct {
	Question      string
	Position      int
	IsMultiAnswer bool
	Answers       []AddSectionQuestionAnswerParams
}

// Represents any type of section params (e.g. quiz/video)
type AddSectionParams interface {
	GetPosition() int
}

type AddVideoSectionParams struct {
	Title      string
	StorageKey uuid.UUID
	Position   int
}

func (v *AddVideoSectionParams) GetPosition() int { return v.Position }

type AddQuizSectionParams struct {
	Position  int
	Questions []AddSectionQuestionParams
}

func (q *AddQuizSectionParams) GetPosition() int { return q.Position }

type AddCourseParams struct {
	Title             string
	Description       string
	CompletionTitle   string
	CompletionMessage string
	Materials         []AddMaterialParams
	Sections          []AddSectionParams
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

// Required so the e2e tests can unmarshal the CourseSection interface that exists
// within the Course struct
func (c *Course) UnmarshalJSON(data []byte) error {
	type alias Course

	// Create an alias of Course with Sections overidden so it unmarshals sections
	// into a json.RawMessage rather than a CourseSection interface
	var raw struct {
		alias
		Sections []json.RawMessage `json:"sections"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Now unmarshal the CourseSection interface manually, depending on the type of each section
	*c = Course(raw.alias)
	c.Sections = make([]CourseSection, 0, len(raw.Sections))
	for _, s := range raw.Sections {
		var typeChecker struct {
			Type SectionType `json:"type"`
		}
		if err := json.Unmarshal(s, &typeChecker); err != nil {
			return fmt.Errorf("missing or invalid section type: %w", err)
		}
		switch typeChecker.Type {
		case SectionTypeVideo:
			var v VideoSection
			if err := json.Unmarshal(s, &v); err != nil {
				return err
			}
			c.Sections = append(c.Sections, &v)
		case SectionTypeQuiz:
			var q QuizSection
			if err := json.Unmarshal(s, &q); err != nil {
				return err
			}
			c.Sections = append(c.Sections, &q)
		default:
			return fmt.Errorf("unknown section type %q", typeChecker.Type)
		}
	}
	return nil
}

type CourseOverview struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
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
