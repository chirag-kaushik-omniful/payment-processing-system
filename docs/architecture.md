# Payment Platform Architecture

See the root [.md](../.md) blueprint for the full handbook.

## Web UI

Vite + React app in [`web/`](../web/):

```bash
make web-install
make web-dev   # http://localhost:5173
```

Proxies `/api` to the gateway. CORS is enabled on api-gateway for `localhost:5173`.

## Local development

```bash
make docker-up
make migrate-up
make build
```

## Service ports

| Service | HTTP Port | Documentation |
|---------|-----------|---------------|
| api-gateway | 8000 | [README](../services/api-gateway/README.md) |
| auth-service | 8001 | [README](../services/auth-service/README.md) |
| payment-service | 8002 | [README](../services/payment-service/README.md) |
| provider-service | 8003 | [README](../services/provider-service/README.md) |
| wallet-service | 8004 | [README](../services/wallet-service/README.md) |
| ledger-service | 8005 | [README](../services/ledger-service/README.md) |
| fraud-service | 8006 | [README](../services/fraud-service/README.md) |
| notification-service | 8007 | [README](../services/notification-service/README.md) |
| webhook-service | 8008 | [README](../services/webhook-service/README.md) |
| reconciliation-service | 8009 | [README](../services/reconciliation-service/README.md) |
| analytics-service | 8010 | [README](../services/analytics-service/README.md) |
