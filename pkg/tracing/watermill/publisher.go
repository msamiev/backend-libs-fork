package watermill

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/tracing"
)

const publisherTracerName = "watermill/publisher"

type PublisherDecorator struct {
	publisher         message.Publisher
	publisherName     string
	spanAttributes    []attribute.KeyValue
	otelTraceProvider trace.TracerProvider
}

func NewPublisherDecorator(pub message.Publisher, options ...PublisherOptionFunc) message.Publisher {
	return NewNamedPublisherDecorator(structName(pub), pub, options...)
}

func NewNamedPublisherDecorator(name string, pub message.Publisher, options ...PublisherOptionFunc) message.Publisher {
	traceProvider, _ := tracing.New(context.Background(), tracing.Config{})
	config := PublisherDecorator{
		otelTraceProvider: traceProvider,
	}
	for _, opt := range options {
		opt(&config)
	}
	return &PublisherDecorator{
		publisher:         pub,
		publisherName:     name,
		spanAttributes:    config.spanAttributes,
		otelTraceProvider: config.otelTraceProvider,
	}
}

type TracerSpan struct {
	ctx  context.Context
	span trace.Span
}

func (pd *PublisherDecorator) Publish(topic string, messages ...*message.Message) error {
	if len(messages) == 0 {
		return nil
	}

	var spans = make([]*TracerSpan, 0)
	for i := range messages {
		spans = append(spans, newTracerSpan(topic, messages[i], pd))
	}

	err := pd.publisher.Publish(topic, messages...)
	if err != nil {
		for i := range spans {
			spans[i].span.RecordError(err)
		}
	}

	for i := range spans {
		spans[i].span.End()
	}
	return err
}

func (pd *PublisherDecorator) Close() error {
	return pd.publisher.Close()
}

func newTracerSpan(topic string, msg *message.Message, config *PublisherDecorator) *TracerSpan {
	msgCtx := msg.Context()
	spanName := message.PublisherNameFromCtx(msgCtx)
	if spanName == "" {
		spanName = config.publisherName
	}

	res := new(TracerSpan)
	res.ctx, res.span = config.otelTraceProvider.Tracer(publisherTracerName).
		Start(msgCtx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	msg.SetContext(res.ctx)
	otel.GetTextMapPropagator().Inject(res.ctx, propagation.MapCarrier(msg.Metadata))

	spanAttributes := []attribute.KeyValue{
		semconv.MessagingDestinationName(topic),
		semconv.MessagingOperationProcess,
		semconv.MessageIDKey.String(msg.UUID),
	}
	spanAttributes = append(spanAttributes, config.spanAttributes...)
	res.span.SetAttributes(spanAttributes...)
	return res
}

func structName(v interface{}) string {
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	s := fmt.Sprintf("%T", v)
	// trim the pointer marker, if any
	return strings.TrimLeft(s, "*")
}

type PublisherOptionFunc func(*PublisherDecorator)

func WithPublisherSpanAttributes(val ...attribute.KeyValue) PublisherOptionFunc {
	return func(config *PublisherDecorator) {
		config.spanAttributes = val
	}
}

func WithPublisherTraceProvider(val trace.TracerProvider) PublisherOptionFunc {
	return func(config *PublisherDecorator) {
		config.otelTraceProvider = val
	}
}
