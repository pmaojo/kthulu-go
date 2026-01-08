package observability

import (
	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/core/metrics"
	"go.opentelemetry.io/otel/metric"
)

// MeterProvider represents metric provider interface.
type MeterProvider interface {
	Meter(name string, opts ...metric.MeterOption) metric.Meter
}

// NewMetricsProvider creates metrics provider based on configuration.
func NewMetricsProvider(cfg *core.Config) (*metrics.PrometheusMetrics, error) {
	// Currently only Prometheus exporter is supported.
	return metrics.NewPrometheusMetrics()
}
