package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vbncursed/vkr/issue-service/internal/http/dto"
	issvc "github.com/vbncursed/vkr/issue-service/internal/service"
)

// JWKS — отдать набор публичных ключей эмитента
// @Summary     JWKS набор ключей
// @Tags        keys
// @Produce     json
// @Success     200 {object} dto.JWKSet
// @Failure     500 {object} APIError
// @Router      /.well-known/keys [get]
func JWKS(svc *issvc.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		keys, err := svc.ListIssuerKeys(c.Request().Context())
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		return writeJSON(c, http.StatusOK, dto.FromIssuerKeys(keys))
	}
}
