# Payment Processing Platform

Production-grade, event-driven payment platform inspired by Stripe/Razorpay. Built in Go with clean architecture, Kafka sagas, Redis caching, and PostgreSQL persistence.

## Documentation

| Doc | Description |
|-----|-------------|
| [.md](.md) | Full platform blueprint |
| [workflow.md](workflow.md) | End-to-end workflow diagrams |
| [docs/ENHANCEMENTS.md](docs/ENHANCEMENTS.md) | **Implemented enhancements** (refund saga, holds, admin, etc.) |
| [docs/architecture.md](docs/architecture.md) | Local development |
| [docs/openapi.yaml](docs/openapi.yaml) | OpenAPI 3 gateway spec |

## Architecture

| Port | Service | README |
|------|---------|--------|
| 8000 | API Gateway — JWT, rate limit, circuit breaker, RBAC | [services/api-gateway/README.md](services/api-gateway/README.md) |
| 8001 | Auth Service | [services/auth-service/README.md](services/auth-service/README.md) |
| 8002 | Payment Service — Saga, outbox, holds, refunds | [services/payment-service/README.md](services/payment-service/README.md) |
| 8003 | Provider Service — Stripe, Razorpay, PayPal | [services/provider-service/README.md](services/provider-service/README.md) |
| 8004 | Wallet Service | [services/wallet-service/README.md](services/wallet-service/README.md) |
| 8005 | Ledger Service — Export API, compensation | [services/ledger-service/README.md](services/ledger-service/README.md) |
| 8006 | Fraud Service — Scoring, blocklists | [services/fraud-service/README.md](services/fraud-service/README.md) |
| 8007 | Notification Service — WS, email, SMS | [services/notification-service/README.md](services/notification-service/README.md) |
| 8008 | Webhook Service — HMAC + Kafka | [services/webhook-service/README.md](services/webhook-service/README.md) |
| 8009 | Reconciliation Service — Auto reconcile | [services/reconciliation-service/README.md](services/reconciliation-service/README.md) |
| 8010 | Analytics Service | [services/analytics-service/README.md](services/analytics-service/README.md) |
| 8011 | Admin Service | [services/admin-service/README.md](services/admin-service/README.md) |
| 8012 | Subscription Service | [services/subscription-service/README.md](services/subscription-service/README.md) |

## Web UI (Vite + React)

```bash
cd web && npm install && npm run dev
# http://localhost:5173 — proxies API to gateway :8000
```

See [web/README.md](web/README.md).

## Quick start

```bash
make docker-up
export DATABASE_URL=postgres://payment:payment@localhost:5432/payment_platform?sslmode=disable
make migrate-up
make build

# Core path (3 terminals)
make run-auth
make run-payment
make run-gateway
```

Copy [.env.example](.env.example) to `.env`.

## Enhanced API flow

1. `POST /auth/login` → JWT
2. `POST /payment-methods` → tokenized method
3. `POST /payments` with `Idempotency-Key` (optional `capture_mode: manual` for holds)
4. Saga: fraud → provider → ledger → wallet debit → notification → `payment.completed`
5. `POST /payments/:id/capture` to capture a hold
6. `POST /refunds` (admin role) → refund saga
7. `POST /reconcile/auto` with settlement file data

## Load & integration testing

```bash
k6 run -e BASE_URL=http://localhost:8000 -e ACCESS_TOKEN=<jwt> tests/k6/load_test.js
go test -tags=integration ./tests/integration/...
```

## Observability

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686

Set `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318` for distributed tracing.

## Make targets

| Target | Description |
|--------|-------------|
| `make build` | Build all binaries to `bin/` |
| `make test` | Run unit tests |
| `make docker-up` | Postgres, Redis, Kafka, Prometheus, Grafana, Jaeger |
| `make migrate-up` | Apply SQL migrations |
# payment-processing-system
