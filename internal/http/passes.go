package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/vbncursed/vkr/issue-service/internal/config"
	"github.com/vbncursed/vkr/issue-service/internal/http/dto"
	issvc "github.com/vbncursed/vkr/issue-service/internal/service"
)

// CreatePass — выпуск пропуска
// @Summary     Выпуск пропуска
// @Tags        passes
// @Accept      json
// @Produce     json
// @Param       request body dto.CreatePassRequest true "Create pass"
// @Success     201 {object} dto.CreatePassResponse
// @Failure     400 {object} APIError
// @Failure     500 {object} APIError
// @Failure     503 {object} APIError
// @Router      /passes [post]
func CreatePass(svc *issvc.Service, cfg config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreatePassRequest
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "malformed"})
		}
		// validations (DTO level)
		if err := req.Validate(time.Now().UTC(), cfg.MaxTTL); err != nil {
			switch err {
			case dto.ErrZoneRequired:
				return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "zone_id required"})
			case dto.ErrNbfAfterExp:
				return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "nbf must be before exp"})
			case dto.ErrExpExceedsMaxTTL:
				return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "exp exceeds max ttl"})
			default:
				return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "malformed"})
			}
		}

		res, err := svc.IssuePass(c.Request().Context(), req.ToCommand())
		if err != nil {
			status, apiErr := MapError(err)
			return writeJSON(c, status, apiErr)
		}
		return writeJSON(c, http.StatusCreated, dto.FromIssueResult(res))
	}
}

// RevokePass — отзыв пропуска
// @Summary     Отзыв пропуска
// @Tags        passes
// @Produce     json
// @Param       id  path string true "Pass ID"
// @Success     200 {object} dto.RevokeResponse
// @Failure     404 {object} APIError
// @Failure     409 {object} APIError
// @Failure     500 {object} APIError
// @Router      /passes/{id}/revoke [post]
func RevokePass(svc *issvc.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "id"})
		}
		if err := svc.RevokePass(c.Request().Context(), id); err != nil {
			status, apiErr := MapError(err)
			return writeJSON(c, status, apiErr)
		}
		return writeJSON(c, http.StatusOK, dto.RevokeResponseOK(id))
	}
}

// ApprovePass — выдаёт pickup-token для забора payload
// @Summary     Сгенерировать pickup-token
// @Tags        pickup
// @Produce     json
// @Param       id  path string true "Pass ID"
// @Success     200 {object} dto.ApproveResponse
// @Failure     404 {object} APIError
// @Failure     409 {object} APIError
// @Failure     500 {object} APIError
// @Router      /passes/{id}/approve [post]
func ApprovePass(svc *issvc.Service, cfg config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "id"})
		}
		res, err := svc.ApprovePass(c.Request().Context(), id, 1*time.Hour)
		if err != nil {
			status, apiErr := MapError(err)
			return writeJSON(c, status, apiErr)
		}
		return writeJSON(c, http.StatusOK, dto.FromApproveResult(id, res))
	}
}

// Pickup — вернуть payload по действующему pickup-токену
// @Summary     Получить payload по pickup-token
// @Tags        pickup
// @Accept      json
// @Produce     json
// @Param       request body dto.PickupRequest true "Pickup"
// @Success     200 {object} dto.PickupResponse
// @Failure     400 {object} APIError
// @Failure     500 {object} APIError
// @Router      /pickup [post]
func Pickup(svc *issvc.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.PickupRequest
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "malformed"})
		}
		if err := req.Validate(); err != nil {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "token required"})
		}
		tok := strings.TrimSpace(req.Token)
		res, err := svc.Pickup(c.Request().Context(), tok)
		if err != nil {
			status, apiErr := MapError(err)
			return writeJSON(c, status, apiErr)
		}
		return writeJSON(c, http.StatusOK, dto.FromPickupResult(res))
	}
}
