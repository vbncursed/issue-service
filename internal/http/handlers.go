package http

import (
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type HealthzResponse struct {
	Status string `json:"status"`
}
type ReadyzResponse struct {
	Status string `json:"status"`
}

// Healthz liveness.
// @Summary     Liveness probe
// @Tags        meta
// @Produce     json
// @Success     200 {object} HealthzResponse
// @Router      /healthz [get]
func Healthz(c echo.Context) error {
	return writeJSON(c, http.StatusOK, HealthzResponse{Status: "ok"})
}

type poolPinger interface {
	Ping(ctx context.Context) error
}

// Readyz readiness (DB ping).
// @Summary     Readiness probe
// @Tags        meta
// @Produce     json
// @Success     200 {object} ReadyzResponse
// @Failure     503 {object} APIError
// @Router      /readyz [get]
func Readyz(pool poolPinger) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 800*time.Millisecond)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return writeJSON(c, http.StatusServiceUnavailable, APIError{Code: "db_not_ready", Message: "db not ready"})
		}
		return writeJSON(c, http.StatusOK, ReadyzResponse{Status: "ready"})
	}
}

// StrictJSONBinder запрещает неизвестные поля
type StrictJSONBinder struct{}

func (StrictJSONBinder) Bind(i interface{}, c echo.Context) error {
	if ct := c.Request().Header.Get(echo.HeaderContentType); ct != "" {
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil || mt != echo.MIMEApplicationJSON {
			return echo.ErrUnsupportedMediaType
		}
	}
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(i); err != nil {
		return err
	}
	return nil
}
