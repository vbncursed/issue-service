package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vbncursed/vkr/issue-service/internal/config"
)

type DBPinger interface{ Ping(ctx echo.Context) error }

func Router(pool *pgxpool.Pool, cfg config.Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Binder = StrictJSONBinder{}
	e.HTTPErrorHandler = DefaultHTTPErrorHandler

	// Swagger UI (включается флагом ENABLE_SWAGGER=1)
	if cfg.EnableSwagger {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	v1 := e.Group("/api/v1")
	v1.GET("/healthz", Healthz)
	v1.GET("/readyz", Readyz(pool))

	// Business endpoints
	v1.POST("/passes", CreatePass(pool, cfg))
	v1.POST("/passes/:id/revoke", RevokePass(pool))
	v1.POST("/passes/:id/approve", ApprovePass(pool, cfg))
	v1.POST("/pickup", Pickup(pool))

	// JWKS
	e.GET("/.well-known/keys", JWKS(pool))

	return e
}
