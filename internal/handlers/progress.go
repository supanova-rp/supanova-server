package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

const progressResource = "user progress"

type GetProgressParams struct {
	CourseID string `json:"courseId" validate:"required"`
}

type UpdateProgressParams struct {
	CourseID  string `json:"courseId" validate:"required"`
	SectionID string `json:"sectionId" validate:"required"`
}

func (h *Handlers) GetProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params GetProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sqlcParams := sqlc.GetProgressParams{
		UserID:   userID,
		CourseID: courseID,
	}

	progress, err := h.Progress.GetProgress(e.Request().Context(), sqlcParams)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(progressResource))
		}

		return internalError(ctx, errors.Getting(progressResource), err, slog.String("id", params.CourseID))
	}

	return e.JSON(http.StatusOK, progress)
}

func (h *Handlers) UpdateProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params UpdateProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sectionID, err := utils.PGUUIDFrom(params.SectionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sqlcParams := sqlc.UpdateProgressParams{
		UserID:    userID,
		CourseID:  courseID,
		SectionID: sectionID,
	}

	err = h.Progress.UpdateProgress(ctx, sqlcParams)
	if err != nil {
		return internalError(ctx, errors.Updating(progressResource), err,
			slog.String("courseId", params.CourseID),
			slog.String("sectionId", params.SectionID))
	}

	return e.NoContent(http.StatusOK)
}

// has completed course query:
// -------------------------------------
// 'SELECT completed_course FROM userprogress WHERE user_id = $1 AND course_id = $2', [userId, courseId],

// user progress update query:
// -------------------------------------
// If there is no existing userprogress (shouldn't happen since user should have some progress already)
// then insert new row with empty completed_section_ids
//       await t.none(
//         `INSERT INTO userprogress (user_id, course_id, completed_section_ids, completed_course)
//        VALUES ($1, $2, ARRAY[]::uuid[], TRUE)
//        ON CONFLICT (user_id, course_id)
//        DO UPDATE SET completed_course = TRUE`,
//         [userId, courseId],
//       );
//     })

func (h *Handlers) SetCourseCompleted(e echo.Context) error {
	// TODO:
	// 1. getUserID from context and parse
	// 2. parse courseId & courseName params
	// 3. check if user has previously completed course
	//   - if yes: 
	//       1. send back 200 response with courseId, completed_course: true
	//   - if no:
	//       2. get user name & email from DB -> create new sqlc query
	//       3. send completion e-mail using email service
	//       4. update user progress (see query below)
	//       5. send back 200 response with courseId, completed_course: true
}
