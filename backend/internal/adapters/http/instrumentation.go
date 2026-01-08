package adapterhttp

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	httpTracer     trace.Tracer
	httpMeter      metric.Meter
	handlerCounter metric.Int64Counter
)

func init() {
	httpTracer = otel.Tracer("adapterhttp")
	httpMeter = otel.Meter("adapterhttp")
	handlerCounter, _ = httpMeter.Int64Counter("http.handler.requests")
}

// instrumentHandler wraps an http.HandlerFunc with tracing span and metrics counter.
func instrumentHandler(name string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := httpTracer.Start(r.Context(), name)
		defer span.End()

		handlerCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", name)))
		h(w, r.WithContext(ctx))
	}
}
