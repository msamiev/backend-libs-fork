package server

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/zap"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/http/server/validator"
)

func Setup(ctx context.Context, c *di.Container, spec *openapi3.T) {
	di.Set(c, di.OptInit(func() (*Readiness, error) {
		return new(Readiness), nil
	}))
	di.Set(c, di.OptInit(func() (*echo.Echo, error) {
		const (
			swaggerSpecPath = "/swagger/swagger.json"
			swaggerUIPath   = "/swagger/*"
		)
		var (
			name = di.GetNamed[string](c, config.AppName)
			p    = prometheus.NewPrometheus(name, nil)
			e    = echo.New()
			log  = di.Get[*zap.Logger](c)
			//tracerProvider = di.Get[trace.TracerProvider](c)
		)
		p.RequestCounterURLLabelMappingFunc = func(c echo.Context) string {
			p := c.Path() // contains route path ala `/users/:id`
			if p != "" {
				return p
			}
			// https://github.com/labstack/echo-contrib/blob/master/prometheus/prometheus.go#L201
			return "unknown"
		}
		e.Logger = NewEchoZapLogger(log)
		e.HideBanner = true
		//e.Use(middleware.CORS(), p.HandlerFunc)
		//e.Use(otelecho.Middleware(name, otelecho.WithTracerProvider(tracerProvider)))
		e.Use(middleware.Recover())
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			//ctx := c.Request().Context()
			//trace.SpanFromContext(ctx).RecordError(err)

			e.DefaultHTTPErrorHandler(err, c)
		}
		if spec != nil {
			e.Use(validator.NewMiddlewareFunc(spec)) //nolint:contextcheck // uses the context from the request
		}

		e.Server.BaseContext = func(_ net.Listener) context.Context { return ctx }

		e.GET(swaggerSpecPath, func(c echo.Context) error {
			return c.JSON(http.StatusOK, spec)
		})
		e.GET(swaggerUIPath, echoSwagger.EchoWrapHandler(func(conf *echoSwagger.Config) {
			conf.URLs = []string{swaggerSpecPath}
		}))

		return e, nil
	}), di.OptDeinit(func(e *echo.Echo) error {
		_ = e.Close()
		return nil
	}))
}
