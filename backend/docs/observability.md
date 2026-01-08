# Observability

The backend exposes runtime metrics and traces through OpenTelemetry.

## Viewing Prometheus metrics

When the service is running, metrics are available on the `/metrics` endpoint:

```bash
curl http://localhost:8080/metrics
```

These metrics can be scraped by Prometheus and viewed with tools like Grafana.

## Viewing traces

Traces are exported using the exporter configured in `OBSERVABILITY_TRACE_EXPORTER`.
Supported values are `stdout` and `jaeger`. For example, to view traces in Jaeger
run the Jaeger collector and set:

```bash
export OBSERVABILITY_TRACE_EXPORTER=jaeger
```

Then start the service and open the Jaeger UI to explore spans produced by the
API handlers and use cases.

