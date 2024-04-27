package nats

import (
	"time"

	nc "github.com/nats-io/nats.go"
)

type ConnectionConfig struct {
	RetryOnFailedConnect bool
	ConnectionTimeout    time.Duration
	ReconnectWait        time.Duration
}

func defaultConnectionConfig() ConnectionConfig {
	const (
		defaultTimeout       = 30 * time.Second
		defaultReconnectWait = 1 * time.Second
	)

	return ConnectionConfig{
		RetryOnFailedConnect: true,
		ConnectionTimeout:    defaultTimeout,
		ReconnectWait:        defaultReconnectWait,
	}
}

func NewConnection(url string, opts ...ConnOptionFunc) (ncc *nc.Conn, err error) {
	config := defaultConnectionConfig()
	for _, opt := range opts {
		opt(&config)
	}
	return nc.Connect(
		url,
		nc.RetryOnFailedConnect(config.RetryOnFailedConnect),
		nc.Timeout(config.ConnectionTimeout),
		nc.ReconnectWait(config.ReconnectWait),
	)
}

type ConnOptionFunc = func(config *ConnectionConfig)

func WithRetryOnFailedConnect(val bool) ConnOptionFunc {
	return func(config *ConnectionConfig) {
		config.RetryOnFailedConnect = val
	}
}

func WithConnectionTimeout(val time.Duration) ConnOptionFunc {
	return func(config *ConnectionConfig) {
		config.ConnectionTimeout = val
	}
}

func WithReconnectWait(val time.Duration) ConnOptionFunc {
	return func(config *ConnectionConfig) {
		config.ReconnectWait = val
	}
}
