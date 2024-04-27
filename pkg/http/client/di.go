package client

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

func Setup(c *di.Container) {
	di.Set(c, di.OptInit(func() (*http.Client, error) {
		return NewClient(
			di.GetNamed[string](c, config.AppName),
			WithTraceProvider(di.Get[trace.TracerProvider](c)),
		)
	}))
}
