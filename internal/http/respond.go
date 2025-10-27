package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func writeJSON(c echo.Context, status int, v any) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.JSON(status, v)
}

func DefaultHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	if he, ok := err.(*echo.HTTPError); ok {
		_ = writeJSON(c, he.Code, map[string]any{
			"code":    http.StatusText(he.Code),
			"message": he.Message,
		})
		return
	}
	_ = writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "internal error"})
}
