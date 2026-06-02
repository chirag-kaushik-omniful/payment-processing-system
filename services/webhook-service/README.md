# Webhook Service

Receives and verifies provider callback webhooks with HMAC signature validation and idempotent processing via Redis.

| Property | Value |
|----------|-------|
| Port | `8008` |
| Stack | Gin, Redis |
| Persistence | Redis (idempotency only) |

## Responsibilities

- Accept provider webhook callbacks (`POST /webhooks/:provider`)
- HMAC-SHA256 signature verification
- Idempotent event processing (replay-safe)
- Replay attack prevention via Redis deduplication

## API Routes

| Method | Path | Headers | Description |
|--------|------|---------|-------------|
| POST | `/webhooks/:provider` | `X-Signature` or `Stripe-Signature` | Receive provider webhook |
| GET | `/health` | — | Health check |
| GET | `/metrics` | — | Prometheus metrics |

Supported providers: `stripe`, `razorpay`

## Workflow

```mermaid
sequenceDiagram
    participant Provider as External_Provider
    participant Webhook as Webhook_Service
    participant Redis

    Provider->>Webhook: POST /webhooks/stripe + signature
    Webhook->>Webhook: HMAC-SHA256 verify body
    alt invalid signature
        Webhook-->>Provider: 401 Unauthorized
    else valid
        Webhook->>Redis: GET idempotency:webhook:stripe:{event_id}
        alt already processed
            Webhook-->>Provider: 200 OK (replay=true)
        else new
            Webhook->>Redis: SETNX lock
            Webhook->>Redis: SET response cache
            Webhook-->>Provider: 200 OK {processed: true}
        end
    end
```

```mermaid
flowchart TD
    ReceiveWebhook[POST_webhooks_provider] --> VerifyHMAC{HMAC_SHA256_valid}
    VerifyHMAC -->|no| Reject401[401_Unauthorized]
    VerifyHMAC -->|yes| IdempotencyGet{Redis_GET_event_id}
    IdempotencyGet -->|cached| ReturnReplay[200_replay_true]
    IdempotencyGet -->|new| TryLock{Redis_SETNX_lock}
    TryLock -->|busy| Conflict409[409_Conflict]
    TryLock -->|acquired| ProcessEvent[Process_webhook]
    ProcessEvent --> CacheResponse[Redis_SET_idempotency]
    CacheResponse --> Return200[200_processed]
```

## ERD (Redis)

```mermaid
erDiagram
    WEBHOOK_IDEMPOTENCY {
        string key PK "idempotency:webhook:{provider}:{event_id}"
        json cached_response
        datetime ttl "48h"
    }

    WEBHOOK_LOCK {
        string key PK "idempotency:webhook:{provider}:{event_id}:lock"
        datetime ttl "30s"
    }

    WEBHOOK_PAYLOAD {
        string event_id
        string event_type
        json data
    }

    WEBHOOK_PAYLOAD ||--o| WEBHOOK_IDEMPOTENCY : deduplicated_by
```

## Security

| Mechanism | Detail |
|-----------|--------|
| HMAC verification | `sha256=HMAC(body, WEBHOOK_SECRET_{PROVIDER})` |
| Idempotency key | `webhook:{provider}:{event_id}` |
| Secret env vars | `WEBHOOK_SECRET_STRIPE`, `WEBHOOK_SECRET_RAZORPAY` |

## Run

```bash
HTTP_PORT=8008 \
WEBHOOK_SECRET_STRIPE=whsec_test_stripe \
WEBHOOK_SECRET_RAZORPAY=whsec_test_razorpay \
go run ./cmd/server
```
