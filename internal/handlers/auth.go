package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const userResource = "user"

type RegisterParams struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *Handlers) Register(e echo.Context) error {
	ctx := e.Request().Context()

	var params RegisterParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	userID, err := h.AuthProvider.CreateUser(ctx, params.Email, params.Password, params.Name)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(userResource), err)
	}

	user, err := h.Auth.RegisterUser(ctx, domain.RegisterParams{
		ID:    userID,
		Name:  params.Name,
		Email: params.Email,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Creating(userResource), err)
	}

	return e.JSON(http.StatusOK, map[string]string{"newUserId": user.ID})
}
