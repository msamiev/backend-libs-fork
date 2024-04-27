package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	env "github.com/caarlos0/env/v6"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

func Setup(ctx context.Context, c *di.Container) {
	di.Set(c, di.OptInit(func() (conf Config, _ error) {
		return conf, env.Parse(&conf) //nolint:gocritic // it's correct evaluation order
	}))
	di.Set(c, di.OptInit(func() (trace.TracerProvider, error) {
		conf := di.Get[Config](c)
		conf.Name = di.GetNamed[string](c, config.AppName)
		conf.Version = di.GetNamed[string](c, config.AppVersion)
		return New(ctx, conf)
	}), di.OptDeinit(func(tp trace.TracerProvider) error {
		shutdownAble, ok := tp.(interface{ Shutdown(context.Context) error })
		if ok {
			return shutdownAble.Shutdown(ctx)
		}

		return nil
	}))
}
