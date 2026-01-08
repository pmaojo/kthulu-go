// @kthulu:core
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	metricsOnce     sync.Once
)

func initMetrics(provider metric.MeterProvider) {
	meter := provider.Meter("kthulu-http")
	requestCounter, _ = meter.Int64Counter("http_server_requests_total")
	requestDuration, _ = meter.Float64Histogram("http_server_request_duration_seconds")
}

// MetricsMiddleware records HTTP request metrics for Prometheus.
func MetricsMiddleware(provider metric.MeterProvider) func(http.Handler) http.Handler {
	metricsOnce.Do(func() { initMetrics(provider) })

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()
			attrs := []attribute.KeyValue{
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
				attribute.Int("status", ww.Status()),
			}

			requestCounter.Add(r.Context(), 1, metric.WithAttributes(attrs...))
			requestDuration.Record(r.Context(), duration, metric.WithAttributes(attrs...))
		})
	}
}
