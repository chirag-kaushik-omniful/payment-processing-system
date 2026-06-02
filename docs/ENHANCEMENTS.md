# Platform Enhancements

Implemented capabilities beyond the original MVP.

## Core payments

| Feature | Location | Description |
|---------|----------|-------------|
| Refund saga | payment + provider | `refund.requested` → `provider.refund` → `refund.completed` → ledger compensation |
| Payment holds | payment-service | `capture_mode: manual` → `AUTHORIZED` + hold; `POST /payments/:id/capture` |
| Multi-currency / FX | `shared/fx` | Static demo rates; converts to base currency on create |
| Tokenized payment methods | payment-service | `POST /payment-methods` stores provider token (no PAN/CVV) |
| PayPal provider | provider-service | Factory adapter alongside Stripe/Razorpay |
| Merchant list API | api-gateway | `GET /merchant/payments` |
| Disputes | payment-service | `POST /disputes`; status `DISPUTED` |

## Saga & events

| Feature | Description |
|---------|-------------|
| `payment.completed` | Published after ledger + notification |
| `payment.processing` | Published when saga starts |
| Ledger compensation | `ledger.compensation` on refund |
| Event versioning | `shared/events` envelope (`schema_version: 1.0`) |
| Wallet debit on success | `wallet.debited` after ledger write |

## Reliability

| Feature | Location |
|---------|----------|
| Kafka retry/DLQ | `shared/kafka` `NewConsumerWithRetry`, `RetryPublisher` |
| Circuit breaker | api-gateway → payment (`shared/circuitbreaker`) |
| Webhook → saga | webhook publishes `webhook.received`; payment updates status |
| k6 load tests | `tests/k6/load_test.js` |
| Integration test stub | `tests/integration/` (build tag `integration`) |

## Security & compliance

| Feature | Description |
|---------|-------------|
| RBAC on gateway | Refunds + `/admin/*` require `admin` role |
| Audit log | `audit_events` table; recorded on payment actions |
| Admin service | Port 8011: list payments, replay saga, view audit |
| Secrets | Use `WEBHOOK_SECRET_*`, `JWT_SECRET`; see `.env.example` |

## Fraud & risk

| Feature | Description |
|---------|-------------|
| Risk score | 0–100 score on `fraud.validated` |
| Device / geo | `device_id`, `country` on validate event |
| Blocklists | `POST /fraud/blocklist`; Redis `fraud:blocklist:{kind}:{id}` |
| Blocked countries | Demo block list `XX`, `ZZ` |

## Operations

| Feature | Location |
|---------|----------|
| Ledger export | `GET /ledger/export?since=&until=&limit=` |
| Auto reconciliation | `POST /reconcile/auto` fetches ledger export |
| Grafana dashboard | `deployments/docker/grafana/dashboards/` |
| Jaeger tracing | docker-compose OTLP ports 4317/4318 |
| OpenAPI | `docs/openapi.yaml` |
| Helm chart | `deployments/helm/payment-platform/` |

## Product

| Feature | Port | Description |
|---------|------|-------------|
| Email / SMS | 8007 | Env: `NOTIFICATION_EMAIL_ENABLED`, `NOTIFICATION_SMS_ENABLED` |
| Subscriptions | 8012 | CRUD + scheduled `subscription.charge` |
| WebSocket | 8007 | `GET /ws` real-time status |

## Feature flags

Set `FEATURE_<NAME>=true|false`:

- `REFUND_SAGA` (default on unless `false`)
- `PAYMENT_HOLDS`
- `AUTO_RECONCILE`
- `MULTI_PROVIDER`

## New services

| Port | Service |
|------|---------|
| 8011 | admin-service |
| 8012 | subscription-service |

## Migrations

Run after `001_init.sql`:

- `payment-service/migrations/002_enhancements.sql`
- `ledger-service/migrations/002_entry_type.sql`
- `wallet-service/migrations/002_holds.sql`
- `subscription-service/migrations/001_subscriptions.sql`
