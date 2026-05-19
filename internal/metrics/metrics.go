package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "iot",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request latency distribution.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "iot",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total HTTP requests.",
	}, []string{"method", "path", "status"})

	TelemetryIngestTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iot",
		Subsystem: "telemetry",
		Name:      "ingest_total",
		Help:      "Total telemetry messages ingested.",
	})

	TelemetryIngestErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iot",
		Subsystem: "telemetry",
		Name:      "ingest_errors_total",
		Help:      "Total telemetry ingest errors.",
	})

	ActiveDevices = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "iot",
		Subsystem: "devices",
		Name:      "active_total",
		Help:      "Number of devices that have sent data in the last 24h.",
	})

	RabbitMQPublishTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iot",
		Subsystem: "rabbitmq",
		Name:      "publish_total",
		Help:      "Total messages published to RabbitMQ.",
	})

	RabbitMQPublishErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iot",
		Subsystem: "rabbitmq",
		Name:      "publish_errors_total",
		Help:      "Total RabbitMQ publish errors.",
	})
)
