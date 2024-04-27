package nats

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	nc "github.com/nats-io/nats.go"
)

type SubscriberConfig struct {
	SubscribersCount    int
	SubscribeTimeout    time.Duration
	AckWaitTimeout      time.Duration
	AckNackCloseTimeout time.Duration
	Logger              watermill.LoggerAdapter
}

func defaultSubscriberConfig() SubscriberConfig {
	const (
		defaultSubscribersCount    = 20
		defaultSubscribeTimeout    = 5 * time.Second
		defaultAckWaitTimeout      = 30 * time.Second
		defaultAckNackCloseTimeout = 5 * time.Second
	)

	return SubscriberConfig{
		SubscribersCount:    defaultSubscribersCount,
		SubscribeTimeout:    defaultSubscribeTimeout,
		AckWaitTimeout:      defaultAckWaitTimeout,
		AckNackCloseTimeout: defaultAckNackCloseTimeout,
		Logger:              watermill.NopLogger{},
	}
}

func NewSubscriber(ncc *nc.Conn, topic, durableName string, opts ...SubscriberOptionFunc) (*nats.Subscriber, error) {
	config := defaultSubscriberConfig()
	for _, opt := range opts {
		opt(&config)
	}
	return subscriber(ncc, topic, durableName, config)
}

func subscriber(ncc *nc.Conn, topic, durableName string, config SubscriberConfig) (*nats.Subscriber, error) {
	return nats.NewSubscriberWithNatsConn(
		ncc,
		nats.SubscriberSubscriptionConfig{
			Unmarshaler: &nats.NATSMarshaler{},
			JetStream: nats.JetStreamConfig{
				SubscribeOptions: []nc.SubOpt{
					nc.Durable(durableName),
					nc.AckExplicit(),
					nc.DeliverAll(),
					nc.ManualAck(),
				},
			},
			SubscribeTimeout: config.SubscribeTimeout,
			AckWaitTimeout:   config.AckWaitTimeout,
			CloseTimeout:     config.AckNackCloseTimeout,
			QueueGroupPrefix: durableName + "-" + topic,
			SubscribersCount: config.SubscribersCount,
		},
		config.Logger,
	)
}

type SubscriberOptionFunc = func(config *SubscriberConfig)

func WithSubscribersCount(val int) SubscriberOptionFunc {
	return func(config *SubscriberConfig) {
		config.SubscribersCount = val
	}
}

func WithSubscribeTimeout(val time.Duration) SubscriberOptionFunc {
	return func(config *SubscriberConfig) {
		config.SubscribeTimeout = val
	}
}

func WithAckWaitTimeout(val time.Duration) SubscriberOptionFunc {
	return func(config *SubscriberConfig) {
		config.AckWaitTimeout = val
	}
}

func WithAckNackCloseTimeout(val time.Duration) SubscriberOptionFunc {
	return func(config *SubscriberConfig) {
		config.AckNackCloseTimeout = val
	}
}

func WithSubscriberLogger(val watermill.LoggerAdapter) SubscriberOptionFunc {
	return func(config *SubscriberConfig) {
		config.Logger = val
	}
}
