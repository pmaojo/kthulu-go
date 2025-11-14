package usecase

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	ucTracer  trace.Tracer
	ucMeter   metric.Meter
	ucCounter metric.Int64Counter
)

func init() {
	ucTracer = otel.Tracer("usecase")
	ucMeter = otel.Meter("usecase")
	ucCounter, _ = ucMeter.Int64Counter("usecase.calls")
}

// startUseCaseSpan begins an OpenTelemetry span and increments a Prometheus counter.
func startUseCaseSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	ctx, span := ucTracer.Start(ctx, name)
	ucCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("usecase", name)))
	return ctx, span
}
