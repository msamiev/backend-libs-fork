package nats

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	nc "github.com/nats-io/nats.go"
)

type PublisherConfig struct {
	RetryAttempts int
	RetryWait     time.Duration
	Logger        watermill.LoggerAdapter
}

func defaultPublisherConfig() PublisherConfig {
	const (
		defaultRetryAttempts = 3
		defaultRetryWait     = 1 * time.Second
	)
	return PublisherConfig{
		RetryAttempts: defaultRetryAttempts,
		RetryWait:     defaultRetryWait,
		Logger:        watermill.NopLogger{},
	}
}

func NewPublisher(ncc *nc.Conn, opts ...PublisherOptionFunc) (*nats.Publisher, error) {
	config := defaultPublisherConfig()
	for _, opt := range opts {
		opt(&config)
	}
	return publisher(ncc, config)
}

func publisher(ncc *nc.Conn, config PublisherConfig) (*nats.Publisher, error) {
	return nats.NewPublisherWithNatsConn(
		ncc,
		nats.PublisherPublishConfig{
			Marshaler: &nats.NATSMarshaler{},
			JetStream: nats.JetStreamConfig{
				ConnectOptions:   []nc.JSOpt{},
				SubscribeOptions: nil,
				PublishOptions: []nc.PubOpt{
					nc.RetryAttempts(config.RetryAttempts),
					nc.RetryWait(config.RetryWait),
				},
				AckAsync: false,
			},
			SubjectCalculator: nats.DefaultSubjectCalculator,
		},
		config.Logger,
	)
}

type PublisherOptionFunc = func(config *PublisherConfig)

func WithRetryAttempts(val int) PublisherOptionFunc {
	return func(config *PublisherConfig) {
		config.RetryAttempts = val
	}
}

func WithRetryWait(val time.Duration) PublisherOptionFunc {
	return func(config *PublisherConfig) {
		config.RetryWait = val
	}
}

func WithPublisherLogger(val watermill.LoggerAdapter) PublisherOptionFunc {
	return func(config *PublisherConfig) {
		config.Logger = val
	}
}
