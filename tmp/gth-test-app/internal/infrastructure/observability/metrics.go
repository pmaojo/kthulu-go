// @kthulu:infrastructure:metrics
package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests.",
		},
		[]string{"path", "method", "status"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(HTTPRequestDuration)
}
