package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Name    string  `env:"-"`
	Version string  `env:"-"`
	Enable  bool    `env:"OTEL_TRACING_ENABLE" envDefault:"true"`
	URL     string  `env:"OTEL_TRACING_URL"    envDefault:"localhost:4317"`
	Ratio   float64 `env:"OTEL_TRACING_RATIO"  envDefault:"1"`
}

func New(ctx context.Context, config Config) (trace.TracerProvider, error) {
	if !config.Enable {
		return trace.NewNoopTracerProvider(), nil
	}

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.URL),
	)
	if err != nil {
		return nil, err
	}

	traceProvider := sdk.NewTracerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithSampler(sdk.ParentBased(sdk.TraceIDRatioBased(config.Ratio))), // https://opentelemetry.io/docs/instrumentation/go/sampling/
		sdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.Name),
			semconv.ServiceVersionKey.String(config.Version),
			semconv.TelemetrySDKLanguageGo,
		)),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return traceProvider, nil
}
