package nats

import (
	"time"

	nc "github.com/nats-io/nats.go"
)

type StreamConfig struct {
	DeduplicateWindow time.Duration
	Replicas          int
	MaxBytes          int64
}

func defaultStreamConfig() StreamConfig {
	const (
		defaultDeduplicateWindow = 24 * time.Hour
		defaultReplicas          = 1
		defaultMaxBytes          = 1_073_741_824 // default 1 GiB (2^30 or 1<<30)
	)

	return StreamConfig{
		DeduplicateWindow: defaultDeduplicateWindow,
		Replicas:          defaultReplicas,
		MaxBytes:          defaultMaxBytes,
	}
}

func CreateStream(js nc.JetStreamContext, topic string, opts ...StreamOptionFunc) error {
	config := defaultStreamConfig()
	for _, opt := range opts {
		opt(&config)
	}

	_, err := js.StreamInfo(topic)
	if err == nil {
		// We are leaving because the stream already exists
		return nil
	}

	_, err = js.AddStream(&nc.StreamConfig{
		Name:        topic,
		Description: "",
		Subjects:    []string{topic},
		Duplicates:  config.DeduplicateWindow,
		Replicas:    config.Replicas,
		Retention:   nc.LimitsPolicy,
		MaxBytes:    config.MaxBytes,
	})
	return err
}

type StreamOptionFunc = func(config *StreamConfig)

func WithDeduplicateWindow(val time.Duration) StreamOptionFunc {
	return func(config *StreamConfig) {
		config.DeduplicateWindow = val
	}
}

func WithReplicas(val int) StreamOptionFunc {
	return func(config *StreamConfig) {
		config.Replicas = val
	}
}

func WithMaxBytes(val int64) StreamOptionFunc {
	return func(config *StreamConfig) {
		config.MaxBytes = val
	}
}
