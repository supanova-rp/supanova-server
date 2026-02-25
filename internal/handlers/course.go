package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

const (
	courseResource         = "course"
	courseOverviewResource = "course overview"
)

type GetCourseParams struct {
	ID string `json:"courseId" validate:"required"`
}

func (h *Handlers) GetCourse(e echo.Context) error {
	ctx := e.Request().Context()

	var params GetCourseParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFromString(params.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	course, err := h.Course.GetCourse(ctx, courseID)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(courseResource))
		}

		return internalError(ctx, errors.Getting(courseResource), err, slog.String("course_id", params.ID))
	}

	enrolled, err := h.isEnrolled(ctx, utils.UUIDFrom(courseID))
	if err != nil {
		return internalError(ctx, errors.Getting(courseResource), err, slog.String("course_id", params.ID))
	}
	if !enrolled {
		return echo.NewHTTPError(http.StatusForbidden, errors.Forbidden(courseResource))
	}

	return e.JSON(http.StatusOK, course)
}

type AddCourseParams struct {
	Title             string              `json:"title" validate:"required"`
	Description       string              `json:"description" validate:"required"`
	CompletionTitle   string              `json:"completionTitle" validate:"required"`
	CompletionMessage string              `json:"completionMessage" validate:"required"`
	Materials         []AddMaterialParams `json:"materials"`
	Sections          []AddSectionParams  `json:"sections" validate:"dive"`
}

type AddMaterialParams struct {
	ID         string `json:"id" validate:"required,uuid"`
	Name       string `json:"name" validate:"required"`
	StorageKey string `json:"storageKey" validate:"required,uuid"`
	Position   int    `json:"position" validate:"gte=0"`
}

type AddQuizAnswerParams struct {
	Answer          string `json:"answer" validate:"required"`
	IsCorrectAnswer bool   `json:"isCorrectAnswer"`
	Position        int    `json:"position" validate:"gte=0"`
}

type AddQuizQuestionParams struct {
	Question      string                `json:"question" validate:"required"`
	Position      int                   `json:"position" validate:"gte=0"`
	IsMultiAnswer bool                  `json:"isMultiAnswer"`
	Answers       []AddQuizAnswerParams `json:"answers" validate:"required,min=1"`
}

type AddVideoSectionParams struct {
	Type       domain.SectionType `json:"type"`
	Title      string             `json:"title"      validate:"required"`
	StorageKey string             `json:"storageKey" validate:"required,uuid"`
	Position   int                `json:"position"   validate:"gte=0"`
}

type AddQuizSectionParams struct {
	Type      domain.SectionType      `json:"type"`
	Position  int                     `json:"position"  validate:"gte=0"`
	Questions []AddQuizQuestionParams `json:"questions" validate:"required,min=1,dive"`
}

type AddSectionParams struct {
	Video *AddVideoSectionParams `validate:"omitempty"`
	Quiz  *AddQuizSectionParams  `validate:"omitempty"`
}

// Custom UnmarshalJSON function needed to handle unmarshalling a section which could be
// a video or quiz (or other section in future)
func (s *AddSectionParams) UnmarshalJSON(data []byte) error {
	// Just unmarshal the type field first to figure out if it is a video/quiz
	var typeChecker struct {
		Type domain.SectionType `json:"type"`
	}
	if err := json.Unmarshal(data, &typeChecker); err != nil {
		return fmt.Errorf("missing or invalid section type: %w", err)
	}

	switch typeChecker.Type {
	case domain.SectionTypeVideo:
		s.Video = &AddVideoSectionParams{}
		return json.Unmarshal(data, s.Video)
	case domain.SectionTypeQuiz:
		s.Quiz = &AddQuizSectionParams{}
		return json.Unmarshal(data, s.Quiz)
	default:
		return fmt.Errorf("unknown section type %q", typeChecker.Type)
	}
}

// Custom MarshalJSON func needed to be able to marshal AddSectionParams in tests
func (s AddSectionParams) MarshalJSON() ([]byte, error) {
	if s.Video != nil {
		return json.Marshal(s.Video)
	}
	if s.Quiz != nil {
		return json.Marshal(s.Quiz)
	}
	return nil, fmt.Errorf("AddSectionParams: neither Video nor Quiz is set")
}

func (h *Handlers) AddCourse(e echo.Context) error {
	ctx := e.Request().Context()

	var req AddCourseParams
	if err := bindAndValidate(e, &req); err != nil {
		return err
	}

	course, err := h.Course.AddCourse(ctx, addCourseParamsFrom(&req))
	if err != nil {
		return internalError(ctx, errors.Creating(courseResource), err)
	}

	return e.JSON(http.StatusCreated, course)
}

func addCourseParamsFrom(req *AddCourseParams) *domain.AddCourseParams {
	materials := utils.Map(req.Materials, func(m AddMaterialParams) domain.AddMaterialParams {
		return domain.AddMaterialParams{
			ID:         uuid.MustParse(m.ID),
			Name:       m.Name,
			StorageKey: uuid.MustParse(m.StorageKey),
			Position:   m.Position,
		}
	})

	sections := utils.Map(req.Sections, func(s AddSectionParams) domain.AddSectionParams {
		switch {
		case s.Video != nil:
			return &domain.AddVideoSectionParams{
				Title:      s.Video.Title,
				StorageKey: uuid.MustParse(s.Video.StorageKey),
				Position:   s.Video.Position,
			}
		case s.Quiz != nil:
			return &domain.AddQuizSectionParams{
				Position:  s.Quiz.Position,
				Questions: addQuizQuestionParamsFrom(s.Quiz.Questions),
			}
		default:
			return nil
		}
	})

	return &domain.AddCourseParams{
		Title:             req.Title,
		Description:       req.Description,
		CompletionTitle:   req.CompletionTitle,
		CompletionMessage: req.CompletionMessage,
		Materials:         materials,
		Sections:          sections,
	}
}

func addQuizQuestionParamsFrom(questions []AddQuizQuestionParams) []domain.AddSectionQuestionParams {
	return utils.Map(questions, func(q AddQuizQuestionParams) domain.AddSectionQuestionParams {
		return domain.AddSectionQuestionParams{
			Question:      q.Question,
			Position:      q.Position,
			IsMultiAnswer: q.IsMultiAnswer,
			Answers: utils.Map(q.Answers, func(a AddQuizAnswerParams) domain.AddSectionQuestionAnswerParams {
				return domain.AddSectionQuestionAnswerParams{
					Answer:          a.Answer,
					IsCorrectAnswer: a.IsCorrectAnswer,
					Position:        a.Position,
				}
			}),
		}
	})
}

func (h *Handlers) GetCoursesOverview(e echo.Context) error {
	ctx := e.Request().Context()

	overviews, err := h.Course.GetCoursesOverview(ctx)
	if err != nil {
		return internalError(ctx, errors.Getting(courseOverviewResource), err)
	}

	return e.JSON(http.StatusOK, overviews)
}
