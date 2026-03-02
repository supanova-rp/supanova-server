package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const (
	quizStateResource   = "quiz state"
	quizAttemptResource = "quiz attempt"
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
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params SaveQuizAttemptParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	answers, err := json.Marshal(params.Answers)
	if err != nil {
		return internalError(ctx, errors.Creating(quizAttemptResource), err, slog.String("quiz_id", params.QuizID))
	}

	err = h.Quiz.SaveQuizAttempt(ctx, domain.SaveQuizAttemptParams{
		UserID:  userID,
		QuizID:  quizID,
		Answers: answers,
	})
	if err != nil {
		return internalError(ctx, errors.Creating(quizAttemptResource), err, slog.String("quiz_id", params.QuizID))
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
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params SetQuizStateParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
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
		return echo.NewHTTPError(http.StatusBadRequest, "invalid quiz state shape")
	}

	err = h.Quiz.SetQuizState(ctx, domain.SetQuizStateParams{
		UserID:    userID,
		QuizID:    quizID,
		QuizState: quizStateRaw,
	})
	if err != nil {
		return internalError(ctx, errors.Creating(quizStateResource), err, slog.String("quiz_id", params.QuizID))
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) SaveQuizState(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params SaveQuizStateParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	quizAnswers, err := json.Marshal(params.Answers)
	if err != nil {
		return internalError(ctx, errors.Creating(quizStateResource), err, slog.String("quiz_id", params.QuizID))
	}

	err = h.Quiz.UpsertQuizState(ctx, domain.UpsertQuizStateParams{
		UserID:      userID,
		QuizID:      quizID,
		QuizAnswers: quizAnswers,
	})
	if err != nil {
		return internalError(ctx, errors.Creating(quizStateResource), err, slog.String("quiz_id", params.QuizID))
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) GetAllQuizSections(e echo.Context) error {
	ctx := e.Request().Context()

	sections, err := h.Quiz.GetAllQuizSections(ctx)
	if err != nil {
		return internalError(ctx, errors.Getting("quiz sections"), err)
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
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params ResetQuizProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	err = h.Quiz.ResetQuizProgress(ctx, userID, quizID)
	if err != nil {
		return internalError(ctx, errors.Deleting(quizStateResource), err, slog.String("quiz_id", params.QuizID))
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
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params GetQuizStateParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	quizID, err := uuid.Parse(params.QuizID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	state, err := h.Quiz.GetQuizState(ctx, userID, quizID)
	if err != nil {
		return internalError(ctx, errors.Getting(quizStateResource), err, slog.String("quiz_id", params.QuizID))
	}

	if state == nil {
		return e.JSON(http.StatusOK, map[string]any{})
	}

	return e.JSON(http.StatusOK, state)
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
		return internalError(ctx, errors.Getting(quizAttemptResource), err)
	}

	return e.JSON(http.StatusOK, attempts)
}
