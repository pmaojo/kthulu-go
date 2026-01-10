// @kthulu:core
package metrics

import (
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// PrometheusMetrics provides a Prometheus exporter and meter provider
// for collecting application metrics.
type PrometheusMetrics struct {
	// Provider is the OpenTelemetry meter provider configured with Prometheus exporter
	Provider *sdkmetric.MeterProvider
	// Handler exposes the Prometheus scrape endpoint
	Handler http.Handler
}

// NewPrometheusMetrics configures OpenTelemetry metrics with a Prometheus exporter.
func NewPrometheusMetrics() (*PrometheusMetrics, error) {
	registry := prom.NewRegistry()
	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	return &PrometheusMetrics{
		Provider: provider,
		Handler:  handler,
	}, nil
}
