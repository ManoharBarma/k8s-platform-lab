# Service Template (Golden Path)

This is the canonical template every banking-platform microservice is scaffolded
from. It demonstrates the standard layered architecture and wires in the shared
platform kit (`pkg/`).

## Layout

```
service-template/
├── cmd/main.go              # entrypoint: config → logging → telemetry → db → http
├── internal/
│   ├── config/              # embeds pkg/config.Base + service-specific fields
│   ├── domain/              # business entities
│   ├── handler/             # HTTP (Gin) transport layer
│   ├── service/             # business logic + Repository interface
│   └── repository/          # persistence (in-memory + Postgres)
├── migrations/              # embedded goose SQL migrations (tmpl_ table prefix)
├── Dockerfile               # multi-stage, distroless non-root
└── go.mod
```

## What you get for free

- **Config** from environment (`pkg/config`), with sensible defaults.
- **Structured logging** (zap) with trace correlation (`pkg/logger`).
- **Telemetry**: OTLP traces + metrics, context propagation (`pkg/telemetry`).
- **HTTP server** with request-ID, logging, recovery, Prometheus and OTel
  middleware, plus graceful shutdown (`pkg/httpserver`).
- **Health probes**: `/healthz`, `/readyz`, `/startupz` (`pkg/health`).
- **Metrics**: `/metrics` for Prometheus (`pkg/metrics`).
- **Persistence**: pgx pool with query tracing + goose migrations, or an
  in-memory fallback when `POSTGRES_ENABLED=false`.

## Endpoints

| Method | Path                     | Description            |
|--------|--------------------------|------------------------|
| GET    | `/healthz`               | Liveness               |
| GET    | `/readyz` / `/startupz`  | Readiness / startup    |
| GET    | `/metrics`               | Prometheus metrics     |
| POST   | `/api/v1/resources`      | Create a resource      |
| GET    | `/api/v1/resources`      | List resources         |
| GET    | `/api/v1/resources/:id`  | Get a resource by ID   |

## Creating a new service from this template

1. Copy `templates/service-template` to `services/<your-service>`.
2. Update the module path in `go.mod` to
   `banking-platform/services/<your-service>` and fix the `replace` path
   (`../../pkg`).
3. Rename the `tmpl_` table prefix to your service's prefix (e.g. `acct_`).
4. Replace the `Resource` domain entity and endpoints with your own.
5. Add the service to the root `go.work`.

## Run locally

```bash
# in-memory mode (no dependencies)
POSTGRES_ENABLED=false OTEL_ENABLED=false go run ./cmd

curl -s localhost:8080/healthz
curl -s -X POST localhost:8080/api/v1/resources -d '{"name":"demo"}'
```
