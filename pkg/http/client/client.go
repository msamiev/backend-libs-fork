package client

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	pph "github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/tracing"
)

type Config struct {
	Timeout            time.Duration
	MaxIdleConnections int
	IdleConnTimeout    time.Duration
	DisableCompression bool

	PrometheusRegisterer prometheus.Registerer
	ConstLabels          map[string]string
	ServiceName          string
	OTELTraceProvider    trace.TracerProvider
}

func defaultConfig() Config {
	const (
		defaultTimeout            = time.Second * 3
		defaultMaxIdleConnections = 100
		defaultIdleConnTimeout    = time.Second * 100
	)

	traceProvider, _ := tracing.New(context.Background(), tracing.Config{})
	return Config{
		Timeout:              defaultTimeout,
		MaxIdleConnections:   defaultMaxIdleConnections,
		IdleConnTimeout:      defaultIdleConnTimeout,
		DisableCompression:   false,
		PrometheusRegisterer: prometheus.DefaultRegisterer,
		ConstLabels:          map[string]string{},
		OTELTraceProvider:    traceProvider,
	}
}

func NewClient(name string, opts ...OptionFunc) (_ *http.Client, err error) {
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       config.MaxIdleConnections,
			IdleConnTimeout:    config.IdleConnTimeout,
			DisableCompression: config.DisableCompression,
		},
		Timeout: config.Timeout,
	}

	if config.ServiceName != "" {
		config.ConstLabels["service"] = config.ServiceName
	}

	if httpClient, err = instrumentClientWithConstLabels(name, httpClient, config); err != nil {
		return nil, err
	}

	httpClient.Transport = otelhttp.NewTransport(
		httpClient.Transport,
		otelhttp.WithTracerProvider(config.OTELTraceProvider),
	)

	return httpClient, nil
}

func instrumentClientWithConstLabels(name string, c *http.Client, config Config) (*http.Client, error) {
	const subsystemHTTPOutgoing = "http_outgoing"

	collector := &outgoingInstrumentation{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   name,
				Subsystem:   subsystemHTTPOutgoing,
				Name:        "requests_total",
				Help:        "A counter for outgoing requests from the wrapped client.",
				ConstLabels: config.ConstLabels,
			},
			[]string{"code", "method"},
		),
		errRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   name,
				Subsystem:   subsystemHTTPOutgoing,
				Name:        "error_requests_total",
				Help:        "A counter for outgoing requests with errors.",
				ConstLabels: config.ConstLabels,
			},
			[]string{},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   name,
				Subsystem:   subsystemHTTPOutgoing,
				Name:        "request_duration_histogram_seconds",
				Help:        "A histogram of outgoing request latencies.",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: config.ConstLabels,
			},
			[]string{"method"},
		),
		dnsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   name,
				Subsystem:   subsystemHTTPOutgoing,
				Name:        "dns_duration_histogram_seconds",
				Help:        "Trace dns latency histogram.",
				Buckets:     []float64{.005, .01, .025, .05},
				ConstLabels: config.ConstLabels,
			},
			[]string{"event"},
		),
		tlsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   name,
				Subsystem:   subsystemHTTPOutgoing,
				Name:        "tls_duration_histogram_seconds",
				Help:        "Trace tls latency histogram.",
				Buckets:     []float64{.05, .1, .25, .5},
				ConstLabels: config.ConstLabels,
			},
			[]string{"event"},
		),
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   name,
			Subsystem:   subsystemHTTPOutgoing,
			Name:        "in_flight_requests",
			Help:        "A gauge of in-flight outgoing requests for the wrapped client.",
			ConstLabels: config.ConstLabels,
		}),
	}

	trace := &pph.InstrumentTrace{
		DNSStart:          func(t float64) { collector.dnsDuration.WithLabelValues("dns_start").Observe(t) },
		DNSDone:           func(t float64) { collector.dnsDuration.WithLabelValues("dns_done").Observe(t) },
		TLSHandshakeStart: func(t float64) { collector.tlsDuration.WithLabelValues("tls_handshake_start").Observe(t) },
		TLSHandshakeDone:  func(t float64) { collector.tlsDuration.WithLabelValues("tls_handshake_done").Observe(t) },
	}

	resultClient := &http.Client{
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
		Timeout:       c.Timeout,
		Transport: pph.InstrumentRoundTripperInFlight(
			collector.inflight, InstrumentRoundTripperErrorCounter(
				collector.errRequests, pph.InstrumentRoundTripperCounter(
					collector.requests, pph.InstrumentRoundTripperTrace(
						trace, pph.InstrumentRoundTripperDuration(collector.duration, c.Transport),
					),
				),
			),
		),
	}

	return resultClient, config.PrometheusRegisterer.Register(collector)
}

type outgoingInstrumentation struct {
	duration    *prometheus.HistogramVec
	requests    *prometheus.CounterVec
	errRequests *prometheus.CounterVec
	dnsDuration *prometheus.HistogramVec
	tlsDuration *prometheus.HistogramVec
	inflight    prometheus.Gauge
}

var _ prometheus.Collector = &outgoingInstrumentation{}

func (i *outgoingInstrumentation) Describe(in chan<- *prometheus.Desc) {
	i.duration.Describe(in)
	i.requests.Describe(in)
	i.errRequests.Describe(in)
	i.dnsDuration.Describe(in)
	i.tlsDuration.Describe(in)
	i.inflight.Describe(in)
}

func (i *outgoingInstrumentation) Collect(in chan<- prometheus.Metric) {
	i.duration.Collect(in)
	i.requests.Collect(in)
	i.errRequests.Collect(in)
	i.dnsDuration.Collect(in)
	i.tlsDuration.Collect(in)
	i.inflight.Collect(in)
}

func InstrumentRoundTripperErrorCounter(counter *prometheus.CounterVec, next http.RoundTripper) pph.RoundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(r)
		if err != nil {
			counter.WithLabelValues().Inc()
		}
		return resp, err
	}
}
