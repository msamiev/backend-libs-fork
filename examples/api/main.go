package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jmoiron/sqlx"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/bootstrap"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/primitives"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/storage/mysql"
)

var specMap = map[string]any{
	"openapi": "3.0.3",
	"info": map[string]any{
		"title":   "Example API",
		"version": "1.0.0",
	},
	"paths": map[string]any{
		"/test": map[string]any{
			"get": map[string]any{
				"responses": map[string]any{
					"200": map[string]any{
						"description": "OK",
					},
				},
			},
		},
	},
}

func main() {
	var (
		c           = di.New()
		ctx, cancel = context.WithCancel(context.Background())
	)

	spec, err := openapi3.NewLoader().LoadFromData(primitives.Must[[]byte](func() ([]byte, error) {
		return primitives.MarshalJSON(specMap)
	}))
	if err != nil {
		di.Get[*zap.Logger](c).Panic("Cannot parse openapi swagger json")
	}

	bootstrap.Setup(ctx, c, "example", "api", spec)
	defer primitives.Must(func() (any, error) { err := c.Release(); return nil, err })

	di.Set(c, di.OptMiddleware(func(e *echo.Echo) (*echo.Echo, error) {
		e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: time.Second,
		}))
		db := di.GetNamed[*sqlx.DB](c, mysql.MainMaster)
		e.GET("/test", func(c echo.Context) error {
			headers := c.Request().Header
			logParams := make([]zap.Field, 0, len(headers))
			for k, v := range headers {
				logParams = append(logParams, zap.Strings(k, v))
			}
			c.Logger().Debug("Headers", logParams)

			const distribution = 5
			if rand.Int63() % distribution == 0 {
				code := http.StatusBadRequest
				return echo.NewHTTPError(code, http.StatusText(code))
			}

			if err := db.PingContext(c.Request().Context()); err != nil {
				code := http.StatusInternalServerError
				return echo.NewHTTPError(code, http.StatusText(code))
			}

			return c.JSON(http.StatusOK, map[string]any{"ok": true})
		})

		return e, nil
	}))

	var (
		logger          = di.Get[*zap.Logger](c)
		introspectionFn = di.Get[bootstrap.Introspection](c)
		preShutdownFn   = di.Get[bootstrap.PreShutdown](c)
		srv             = di.Get[*echo.Echo](c)
	)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(introspectionFn)
	g.Go(func() (err error) {
		preShutdownFn(ctx)
		err = srv.Shutdown(ctx)
		cancel()
		return err
	})
	g.Go(func() (err error) {
		if err = srv.Start(":8080"); errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	})
	if err := g.Wait(); err != nil {
		logger.Panic("Fails", zap.Error(err))
	}
}
