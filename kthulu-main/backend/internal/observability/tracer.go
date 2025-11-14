package observability

import (
	"backend/core"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider represents an OpenTelemetry tracer provider interface.
type TracerProvider interface {
	Tracer(name string, opts ...trace.TracerOption) trace.Tracer
}

// NewTracerProvider configures an OpenTelemetry tracer provider based on configuration.
func NewTracerProvider(cfg *core.Config) (*sdktrace.TracerProvider, error) {
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Observability.TraceSampleRate))

	var (
		exp sdktrace.SpanExporter
		err error
	)

	switch cfg.Observability.TraceExporter {
	case "stdout":
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "jaeger":
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint())
	default:
		return nil, fmt.Errorf("unsupported trace exporter: %s", cfg.Observability.TraceExporter)
	}
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exp),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
