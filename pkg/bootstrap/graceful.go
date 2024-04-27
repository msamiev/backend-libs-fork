package bootstrap

import (
	"context"
	"errors"
	"net"
	"net/http"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/http/server"
)

type (
	Introspection func() error
	PreShutdown   func(context.Context)
)

func GracefulSetup(c *di.Container) {
	di.Set(c, di.OptInit(func() (*http.Server, error) {
		var (
			conf = di.Get[config.Introspection](c)
			ctx  = di.Get[context.Context](c)
		)
		return &http.Server{
			Addr:              conf.Sock,
			BaseContext:       func(net.Listener) context.Context { return ctx },
			ReadHeaderTimeout: conf.Timeout,
		}, nil
	}), di.OptDeinit(func(srv *http.Server) error {
		_ = srv.Close()
		return nil
	}))

	di.Set(c, di.OptInit(func() (Introspection, error) {
		var (
			conf      = di.Get[config.Introspection](c)
			logger    = di.Get[*zap.Logger](c).Sugar()
			readiness = di.Get[*server.Readiness](c)
			srv       = di.Get[*http.Server](c)
		)
		if conf.Sock == "" {
			logger.Warn("Introspection disabled")
			return func() error { return nil }, nil
		}

		return func() (err error) {
			const (
				metricsURL   = "/metrics"
				pprofURL     = "/debug/pprof"
				readinessURL = "/readiness"
			)
			logger.Infof("Serve pprof from %s%s", conf.Sock, pprofURL)

			logger.Infof("Serve metrics from %s%s", conf.Sock, metricsURL)
			http.Handle(metricsURL, promhttp.Handler())

			logger.Infof("Serve readiness probe from %s%s", conf.Sock, readinessURL)
			http.Handle(readinessURL, readiness)

			if err = srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			return err
		}, nil
	}))

	di.Set(c, di.OptInit(func() (PreShutdown, error) {
		var (
			cancel    = di.Get[context.CancelFunc](c)
			conf      = di.Get[config.Introspection](c)
			ctx       = di.Get[context.Context](c)
			logger    = di.Get[*zap.Logger](c).Sugar()
			readiness = di.Get[*server.Readiness](c)
			srv       = di.Get[*http.Server](c)
		)
		return func(externalCtx context.Context) {
			select {
			case <-ctx.Done():
			case <-externalCtx.Done():
				cancel()
			}

			readiness.Shutdown(conf.ShutdownNum)
			readiness.Wait(conf.Timeout)

			if conf.Sock != "" {
				if err := srv.Close(); err != nil {
					logger.Warn("Introspection server err", zap.Error(err))
				}
			}
		}, nil
	}))
}
