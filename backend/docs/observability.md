# Observability Setup

This guide provisions Prometheus, Grafana and Jaeger for local metrics and tracing.

## Prerequisites
- Docker and Docker Compose

## Application configuration

Tracing is configured via environment variables:

- `TRACE_EXPORTER` selects the tracing backend. Supported values are `stdout`
  (default) and `jaeger`.
- `TRACE_SAMPLE_RATE` controls the sampling ratio (defaults to `1`).

When using Jaeger, the exporter reads the collector endpoint from
`OTEL_EXPORTER_JAEGER_ENDPOINT` (default `http://localhost:14268/api/traces`).

Example:

```sh
TRACE_EXPORTER=jaeger \
OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces \
go run cmd/service/main.go
```

## Start the stack
```sh
docker compose -f deploy/monitoring/docker-compose.yml up -d
```

Prometheus will be available at <http://localhost:9090>, Grafana at <http://localhost:3000> and Jaeger at <http://localhost:16686>.

## Configure Grafana
1. Login with `admin`/`admin`.
2. Add a Prometheus data source pointing to `http://prometheus:9090`.
3. Add a Jaeger data source pointing to `http://jaeger:16686`.
4. Import dashboards from `deploy/monitoring/dashboards/`.

Example dashboard showing HTTP metrics:
![HTTP Dashboard](https://grafana.com/static/assets/img/docs/dashboards/metrics-explorer.png)

Jaeger UI displays traces collected from the service:
![Jaeger UI](https://www.jaegertracing.io/img/jaeger-ui.png)

## Scrape configuration
Prometheus is configured to scrape the service at `/metrics` in `deploy/monitoring/prometheus.yml`.

## References
- [Prometheus documentation](https://prometheus.io/docs/introduction/overview/)
- [Grafana documentation](https://grafana.com/docs/)
- [Jaeger documentation](https://www.jaegertracing.io/docs/latest/)
