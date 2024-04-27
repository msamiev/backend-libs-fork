package client

import (
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
)

type OptionFunc = func(config *Config)

func WithTimeout(timeout time.Duration) OptionFunc {
	return func(config *Config) {
		config.Timeout = timeout
	}
}

func WithMaxIdleConnections(count int) OptionFunc {
	return func(config *Config) {
		config.MaxIdleConnections = count
	}
}

func WithIdleConnTimeout(timeout time.Duration) OptionFunc {
	return func(config *Config) {
		config.IdleConnTimeout = timeout
	}
}

func WithDisableCompression(isDisabled bool) OptionFunc {
	return func(config *Config) {
		config.DisableCompression = isDisabled
	}
}

func WithPrometheusRegisterer(registerer prometheus.Registerer) OptionFunc {
	return func(config *Config) {
		config.PrometheusRegisterer = registerer
	}
}

func WithConstLabels(labels map[string]string) OptionFunc {
	return func(config *Config) {
		config.ConstLabels = labels
	}
}

func WithServiceName(serviceName string) OptionFunc {
	return func(config *Config) {
		config.ServiceName = serviceName
	}
}

func WithTraceProvider(provider trace.TracerProvider) OptionFunc {
	return func(config *Config) {
		config.OTELTraceProvider = provider
	}
}
