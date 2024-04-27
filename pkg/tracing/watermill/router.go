package watermill

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/tracing"
)

const routerTracerName = "watermill/router"

type RouterDecorator struct {
	router            *message.Router
	spanAttributes    []attribute.KeyValue
	otelTraceProvider trace.TracerProvider
}

func NewRouterDecorator(router *message.Router, options ...HandlerOptionFunc) *RouterDecorator {
	traceProvider, _ := tracing.New(context.Background(), tracing.Config{})
	config := &RouterDecorator{
		router:            router,
		otelTraceProvider: traceProvider,
	}
	for _, opt := range options {
		opt(config)
	}
	return config
}

func (rd *RouterDecorator) AddHandler(
	handlerName string,
	subscribeTopic string,
	subscriber message.Subscriber,
	publishTopic string,
	publisher message.Publisher,
	handlerFunc message.HandlerFunc,
) *message.Handler {
	return rd.router.AddHandler(
		handlerName,
		subscribeTopic,
		subscriber,
		publishTopic,
		publisher,
		rd.handler(handlerFunc),
	)
}

func (rd *RouterDecorator) handler(callback message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (messages []*message.Message, err error) {
		ctx, span := rd.startTrace(msg)
		msg.SetContext(ctx)
		messages, err = callback(msg)
		if err != nil {
			span.RecordError(err)
		}
		span.End()
		return messages, err
	}
}

func (rd *RouterDecorator) AddNoPublisherHandler(
	handlerName string,
	subscribeTopic string,
	subscriber message.Subscriber,
	handlerFunc message.NoPublishHandlerFunc,
) *message.Handler {
	return rd.router.AddNoPublisherHandler(
		handlerName,
		subscribeTopic,
		subscriber,
		rd.noPublishHandler(handlerFunc),
	)
}

func (rd *RouterDecorator) noPublishHandler(callback message.NoPublishHandlerFunc) message.NoPublishHandlerFunc {
	return func(msg *message.Message) (err error) {
		ctx, span := rd.startTrace(msg)
		msg.SetContext(ctx)
		err = callback(msg)
		if err != nil {
			span.RecordError(err)
		}
		span.End()
		return err
	}
}

func (rd *RouterDecorator) startTrace(msg *message.Message) (ctx context.Context, span trace.Span) {
	spanOptions := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindConsumer),
	}
	spanName := message.SubscriberNameFromCtx(msg.Context())
	propagationCtx := otel.GetTextMapPropagator().Extract(msg.Context(), propagation.MapCarrier(msg.Metadata))

	ctx, span = rd.otelTraceProvider.Tracer(routerTracerName).Start(propagationCtx, spanName, spanOptions...)
	spanAttributes := []attribute.KeyValue{
		semconv.MessagingDestinationName(message.SubscribeTopicFromCtx(ctx)),
		semconv.MessagingOperationReceive,
		semconv.MessageIDKey.String(msg.UUID),
	}
	spanAttributes = append(spanAttributes, rd.spanAttributes...)
	span.SetAttributes(spanAttributes...)

	return ctx, span
}

type HandlerOptionFunc func(*RouterDecorator)

func WithHandlerSpanAttributes(val ...attribute.KeyValue) HandlerOptionFunc {
	return func(config *RouterDecorator) {
		config.spanAttributes = val
	}
}

func WithHandlerTraceProvider(val trace.TracerProvider) HandlerOptionFunc {
	return func(config *RouterDecorator) {
		config.otelTraceProvider = val
	}
}
