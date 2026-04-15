package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

const (
	courseResource               = "course"
	coursesResource              = "courses"
	courseOverviewResource       = "course overview"
	assignedCourseTitlesResource = "assigned course titles"
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
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	course, err := h.Course.GetCourse(ctx, courseID)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return httpError(http.StatusNotFound, errors.NotFound(courseResource), err)
		}

		return httpError(http.StatusInternalServerError, errors.Getting(courseResource), err)
	}

	enrolled, err := h.isEnrolled(ctx, utils.UUIDFrom(courseID))
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(courseResource), err)
	}
	if !enrolled {
		return httpError(http.StatusForbidden, errors.Forbidden(courseResource), nil)
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
		return httpError(http.StatusInternalServerError, errors.Creating(courseResource), err)
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

type EditCourseRequest struct {
	CourseID           string             `json:"edited_course_id" validate:"required,uuid"`
	EditedCourse       EditedCourseFields `json:"edited_course" validate:"required"`
	DeletedSectionIDs  DeletedSectionIDs  `json:"deleted_section_ids_map" validate:"required"`
	DeletedMaterialIDs []string           `json:"deleted_materials_ids" validate:"dive,uuid"`
}

type EditedCourseFields struct {
	Title             string               `json:"title" validate:"required"`
	Description       string               `json:"description" validate:"required"`
	CompletionTitle   string               `json:"completionTitle" validate:"required"`
	CompletionMessage string               `json:"completionMessage" validate:"required"`
	Materials         []EditMaterialParams `json:"materials" validate:"dive"`
	Sections          []EditSectionParams  `json:"sections" validate:"dive"`
}

type EditMaterialParams struct {
	ID         string `json:"id" validate:"required,uuid"`
	Name       string `json:"name" validate:"required"`
	StorageKey string `json:"storageKey" validate:"required,uuid"`
	Position   int    `json:"position" validate:"gte=0"`
}

type DeletedSectionIDs struct {
	VideoSectionIDs []string `json:"videoSectionIds" validate:"dive,uuid"`
	QuizSectionIDs  []string `json:"quizSectionIds" validate:"dive,uuid"`
	QuestionIDs     []string `json:"questionIds" validate:"dive,uuid"`
	AnswerIDs       []string `json:"answerIds" validate:"dive,uuid"`
}

type EditSectionParams struct {
	Video *EditVideoSectionParams
	Quiz  *EditQuizSectionParams
}

func (s *EditSectionParams) UnmarshalJSON(data []byte) error {
	var typeChecker struct {
		Type domain.SectionType `json:"type"`
	}
	if err := json.Unmarshal(data, &typeChecker); err != nil {
		return fmt.Errorf("missing or invalid section type: %w", err)
	}

	switch typeChecker.Type {
	case domain.SectionTypeVideo:
		s.Video = &EditVideoSectionParams{}
		return json.Unmarshal(data, s.Video)
	case domain.SectionTypeQuiz:
		s.Quiz = &EditQuizSectionParams{}
		return json.Unmarshal(data, s.Quiz)
	default:
		return fmt.Errorf("unknown section type %q", typeChecker.Type)
	}
}

// Custom MarshalJSON needed to serialise EditSectionParams correctly in tests
func (s EditSectionParams) MarshalJSON() ([]byte, error) {
	if s.Video != nil {
		return json.Marshal(s.Video)
	}
	if s.Quiz != nil {
		return json.Marshal(s.Quiz)
	}
	return nil, fmt.Errorf("EditSectionParams: neither Video nor Quiz is set")
}

type EditVideoSectionParams struct {
	Type         domain.SectionType `json:"type"`
	ID           string             `json:"id"` // skip validation: ID UUID/unix timestamp if existing/new section
	IsNewSection bool               `json:"isNewSection"`
	Title        string             `json:"title" validate:"required"`
	StorageKey   string             `json:"storageKey" validate:"required,uuid"`
	Position     int                `json:"position" validate:"gte=0"`
}

type EditQuizAnswerParams struct {
	ID              string `json:"id" validate:"required,uuid"`
	Answer          string `json:"answer" validate:"required"`
	IsCorrectAnswer bool   `json:"isCorrectAnswer"`
	Position        int    `json:"position" validate:"gte=0"`
}

type EditQuizQuestionParams struct {
	ID            string                 `json:"id" validate:"required,uuid"`
	Question      string                 `json:"question" validate:"required"`
	Position      int                    `json:"position" validate:"gte=0"`
	IsMultiAnswer bool                   `json:"isMultiAnswer"`
	Answers       []EditQuizAnswerParams `json:"answers" validate:"required,min=1,dive"`
}

type EditQuizSectionParams struct {
	Type         domain.SectionType       `json:"type"`
	ID           string                   `json:"id"` // skip validation: ID UUID/unix timestamp if existing/new section
	IsNewSection bool                     `json:"isNewSection"`
	Position     int                      `json:"position" validate:"gte=0"`
	Questions    []EditQuizQuestionParams `json:"questions" validate:"required,min=1,dive"`
}

func (h *Handlers) EditCourse(e echo.Context) error {
	ctx := e.Request().Context()

	var req EditCourseRequest
	if err := bindAndValidate(e, &req); err != nil {
		return err
	}

	params, err := editCourseParamsFrom(&req)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	course, err := h.Course.EditCourse(ctx, params)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(courseResource), err)
	}

	return e.JSON(http.StatusOK, course)
}

func editCourseParamsFrom(req *EditCourseRequest) (*domain.EditCourseParams, error) {
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		return nil, fmt.Errorf("invalid course ID: %w", err)
	}

	materials := make([]domain.AddMaterialParams, 0, len(req.EditedCourse.Materials))
	for _, m := range req.EditedCourse.Materials {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid material ID: %w", err)
		}
		storageKey, err := uuid.Parse(m.StorageKey)
		if err != nil {
			return nil, fmt.Errorf("invalid material storage key: %w", err)
		}
		materials = append(materials, domain.AddMaterialParams{
			ID:         id,
			Name:       m.Name,
			StorageKey: storageKey,
			Position:   m.Position,
		})
	}

	var newVideoSections []domain.AddVideoSectionParams
	var existingVideoSections []domain.EditVideoSectionParams
	var quizSections []domain.EditQuizSectionParams

	for _, s := range req.EditedCourse.Sections {
		switch {
		case s.Video != nil:
			storageKey, err := uuid.Parse(s.Video.StorageKey)
			if err != nil {
				return nil, fmt.Errorf("invalid video section storage key: %w", err)
			}
			if s.Video.IsNewSection {
				newVideoSections = append(newVideoSections, domain.AddVideoSectionParams{
					Title:      s.Video.Title,
					StorageKey: storageKey,
					Position:   s.Video.Position,
				})
			} else {
				id, err := uuid.Parse(s.Video.ID)
				if err != nil {
					return nil, fmt.Errorf("invalid video section ID: %w", err)
				}
				existingVideoSections = append(existingVideoSections, domain.EditVideoSectionParams{
					ID:         id,
					Title:      s.Video.Title,
					StorageKey: storageKey,
					Position:   s.Video.Position,
				})
			}
		case s.Quiz != nil:
			var quizID uuid.UUID
			if !s.Quiz.IsNewSection {
				quizID, err = uuid.Parse(s.Quiz.ID)
				if err != nil {
					return nil, fmt.Errorf("invalid quiz section ID: %w", err)
				}
			}
			questions, err := editQuizQuestionParamsFrom(s.Quiz.Questions)
			if err != nil {
				return nil, err
			}
			quizSections = append(quizSections, domain.EditQuizSectionParams{
				ID:           quizID,
				IsNewSection: s.Quiz.IsNewSection,
				Position:     s.Quiz.Position,
				Questions:    questions,
			})
		}
	}

	deletedVideoSectionIDs, err := parseUUIDs(req.DeletedSectionIDs.VideoSectionIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted video section ID: %w", err)
	}
	deletedQuizSectionIDs, err := parseUUIDs(req.DeletedSectionIDs.QuizSectionIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted quiz section ID: %w", err)
	}
	deletedQuestionIDs, err := parseUUIDs(req.DeletedSectionIDs.QuestionIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted question ID: %w", err)
	}
	deletedAnswerIDs, err := parseUUIDs(req.DeletedSectionIDs.AnswerIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted answer ID: %w", err)
	}
	deletedMaterialIDs, err := parseUUIDs(req.DeletedMaterialIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted material ID: %w", err)
	}

	return &domain.EditCourseParams{
		CourseID:              courseID,
		Title:                 req.EditedCourse.Title,
		Description:           req.EditedCourse.Description,
		CompletionTitle:       req.EditedCourse.CompletionTitle,
		CompletionMessage:     req.EditedCourse.CompletionMessage,
		Materials:             materials,
		NewVideoSections:      newVideoSections,
		ExistingVideoSections: existingVideoSections,
		QuizSections:          quizSections,
		DeletedSectionIDs: domain.DeletedSectionIDs{
			VideoSectionIDs: deletedVideoSectionIDs,
			QuizSectionIDs:  deletedQuizSectionIDs,
			QuestionIDs:     deletedQuestionIDs,
			AnswerIDs:       deletedAnswerIDs,
		},
		DeletedMaterialIDs: deletedMaterialIDs,
	}, nil
}

func editQuizQuestionParamsFrom(questions []EditQuizQuestionParams) ([]domain.EditQuizQuestionParams, error) {
	result := make([]domain.EditQuizQuestionParams, 0, len(questions))
	for _, q := range questions {
		id, err := uuid.Parse(q.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid quiz question ID: %w", err)
		}
		answers := make([]domain.EditQuizAnswerParams, 0, len(q.Answers))
		for _, a := range q.Answers {
			answerID, err := uuid.Parse(a.ID)
			if err != nil {
				return nil, fmt.Errorf("invalid quiz answer ID: %w", err)
			}
			answers = append(answers, domain.EditQuizAnswerParams{
				ID:              answerID,
				Answer:          a.Answer,
				IsCorrectAnswer: a.IsCorrectAnswer,
				Position:        a.Position,
			})
		}
		result = append(result, domain.EditQuizQuestionParams{
			ID:            id,
			Question:      q.Question,
			Position:      q.Position,
			IsMultiAnswer: q.IsMultiAnswer,
			Answers:       answers,
		})
	}
	return result, nil
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

type DeleteCourseParams struct {
	CourseID string `json:"course_id" validate:"required"`
}

func (h *Handlers) DeleteCourse(e echo.Context) error {
	ctx := e.Request().Context()

	var params DeleteCourseParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	if err := h.Course.DeleteCourse(ctx, courseID); err != nil {
		return httpError(http.StatusInternalServerError, errors.Deleting(courseResource), err)
	}

	return e.JSON(http.StatusOK, params.CourseID)
}

func (h *Handlers) GetCourses(e echo.Context) error {
	ctx := e.Request().Context()

	courses, err := h.Course.GetAllCourses(ctx)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(coursesResource), err)
	}

	return e.JSON(http.StatusOK, courses)
}

func (h *Handlers) GetCoursesOverview(e echo.Context) error {
	ctx := e.Request().Context()

	overviews, err := h.Course.GetCoursesOverview(ctx)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(courseOverviewResource), err)
	}

	return e.JSON(http.StatusOK, overviews)
}

func (h *Handlers) GetAssignedCourseTitles(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	overviews, err := h.Course.GetAssignedCourseTitles(ctx, userID)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(assignedCourseTitlesResource), err)
	}

	return e.JSON(http.StatusOK, overviews)
}
