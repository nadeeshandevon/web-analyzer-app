package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		APILabels,
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	registerDashboardOnce sync.Once
)

func RegisterAPIMetrics() {
	registerDashboardOnce.Do(func() {
		prometheus.MustRegister(httpRequestsTotal)
		prometheus.MustRegister(httpRequestDuration)
	})
}

func RecordHttpRequestTotal(method string, path string, status string) {
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
}

func RecordHttpRequestDuration(method string, path string, duration float64) {
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}
