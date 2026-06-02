# Payment Platform Web UI

Vite + React + TypeScript dashboard for the payment platform API gateway.

## Prerequisites

- Node.js 18+
- API gateway on `http://localhost:8000` (with CORS for `http://localhost:5173`)
- Auth + payment services running

## Development

```bash
cd web
npm install
npm run dev
```

Open http://localhost:5173

Vite proxies `/api/*` → `http://localhost:8000/*` (see `vite.config.ts`).

## Production build

```bash
npm run build
npm run preview
```

Set `VITE_API_BASE` to your gateway URL when not using the dev proxy.

## Features

| Page | APIs |
|------|------|
| Login / Signup | `POST /auth/login`, `/auth/signup` |
| Dashboard | `/health`, `/payments`, `/wallet` |
| Payments | Create, list, get, capture hold, idempotency key |
| Wallet | `GET /wallet` |
| Payment methods | `POST /payment-methods` |
| Admin | List payments, audit, replay saga, refunds |

Admin and refund routes require the `admin` role on your JWT.
# payment-processing-system
