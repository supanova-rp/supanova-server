package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const (
	quizStateResource     = "quiz state"
	quizAttemptResource   = "quiz attempt"
	quizQuestionsResource = "quiz questions"
)

type QuizStateAnswers struct {
	QuestionID        string   `json:"questionID" validate:"required"`
	SelectedAnswerIDs []string `json:"selectedAnswerIDs" validate:"required"`
	Correct           bool     `json:"correct"`
}

type SaveQuizAttemptParams struct {
	QuizID  string             `json:"quizID" validate:"required"`
	Answers []QuizStateAnswers `json:"answers" validate:"required,dive"`
}

func (h *Handlers) SaveQuizAttempt(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params SaveQuizAttemptParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	answers, err := json.Marshal(params.Answers)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(quizAttemptResource), err)
	}

	err = h.Quiz.SaveQuizAttempt(ctx, domain.SaveQuizAttemptParams{
		UserID:  userID,
		QuizID:  quizID,
		Answers: answers,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(quizAttemptResource), err)
	}

	return e.NoContent(http.StatusNoContent)
}

type SaveQuizStateParams struct {
	QuizID  string             `json:"quizID" validate:"required"`
	Answers []QuizStateAnswers `json:"answers" validate:"required,dive"`
}

type SetQuizStateParams struct {
	// TODO: change to quizID in future for consistency once FE is updated
	QuizID string          `json:"quizId" validate:"required"`
	State  json.RawMessage `json:"quizState" validate:"required"`
}

func (h *Handlers) SetQuizState(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params SetQuizStateParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	// Client sends state as a JSON string, e.g. "[[0, 3], [7]]"
	// so we have to unmarshal the string into actual raw json, e.g. [[0, 3], [7]]
	// Usually would fix this on the client side but this endpoint should be deprecated at some point
	// in favour of /quiz/save-state
	quizStateRaw := params.State
	var asString string
	if err := json.Unmarshal(quizStateRaw, &asString); err == nil {
		quizStateRaw = json.RawMessage(asString)
	}

	// Validate the shape of the quizState
	var quizState [][]int
	if err := json.Unmarshal(quizStateRaw, &quizState); err != nil {
		return httpError(http.StatusBadRequest, "invalid quiz state shape", err)
	}

	err = h.Quiz.SetQuizState(ctx, domain.SetQuizStateParams{
		UserID:    userID,
		QuizID:    quizID,
		QuizState: quizStateRaw,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(quizStateResource), err)
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) SaveQuizState(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params SaveQuizStateParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	quizAnswers, err := json.Marshal(params.Answers)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(quizStateResource), err)
	}

	err = h.Quiz.UpsertQuizState(ctx, domain.UpsertQuizStateParams{
		UserID:      userID,
		QuizID:      quizID,
		QuizAnswers: quizAnswers,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(quizStateResource), err)
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) GetAllQuizSections(e echo.Context) error {
	ctx := e.Request().Context()

	sections, err := h.Quiz.GetAllQuizSections(ctx)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting("quiz sections"), err)
	}

	return e.JSON(http.StatusOK, sections)
}

type ResetQuizProgressParams struct {
	QuizID string `json:"quizID" validate:"required"`
}

func (h *Handlers) ResetQuizProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params ResetQuizProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	err = h.Quiz.ResetQuizProgress(ctx, userID, quizID)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Deleting(quizStateResource), err)
	}

	return e.NoContent(http.StatusNoContent)
}

type GetQuizStateParams struct {
	// TODO: change to quizID in future for consistency once FE is updated
	QuizID string `json:"quizId" validate:"required"`
}

func (h *Handlers) GetQuizState(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params GetQuizStateParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	state, err := h.Quiz.GetQuizState(ctx, userID, quizID)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(quizStateResource), err)
	}

	if state == nil {
		return e.JSON(http.StatusOK, map[string]any{})
	}

	return e.JSON(http.StatusOK, state)
}

type GetQuizQuestionsParams struct {
	QuizSectionIDs []string `json:"quizSectionIds" validate:"required"`
}

func (h *Handlers) GetQuizQuestions(e echo.Context) error {
	ctx := e.Request().Context()

	var params GetQuizQuestionsParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	if len(params.QuizSectionIDs) == 0 {
		return e.JSON(http.StatusOK, []domain.QuizQuestionLegacy{})
	}

	sectionIDs := make([]uuid.UUID, 0, len(params.QuizSectionIDs))
	for _, id := range params.QuizSectionIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
		}
		sectionIDs = append(sectionIDs, parsed)
	}

	questions, err := h.Quiz.GetQuizQuestions(ctx, sectionIDs)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(quizQuestionsResource), err)
	}

	return e.JSON(http.StatusOK, questions)
}

type GetQuizAttemptsParams struct {
	UserID string `json:"userID" validate:"required"`
}

func (h *Handlers) GetQuizAttemptsByUserID(e echo.Context) error {
	ctx := e.Request().Context()

	var params GetQuizAttemptsParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	attempts, err := h.Quiz.GetQuizAttemptsByUserID(ctx, params.UserID)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(quizAttemptResource), err)
	}

	return e.JSON(http.StatusOK, attempts)
}
