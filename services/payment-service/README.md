# Payment Service

Core orchestrator for the payment lifecycle. Coordinates the distributed saga, enforces idempotency, and uses the transactional outbox pattern for reliable Kafka publishing.

| Property | Value |
|----------|-------|
| Port | `8002` |
| Stack | Gin, PostgreSQL, Redis, Kafka |
| Persistence | PostgreSQL + Redis |

## Responsibilities

- Payment CRUD and refund initiation
- State machine: `CREATED → PROCESSING → SUCCESS | FAILED → REFUNDED`
- Transactional outbox (DB + event in one transaction)
- Saga coordination via Kafka consumers
- Redis idempotency for duplicate requests

## API Routes

| Method | Path | Headers | Description |
|--------|------|---------|-------------|
| POST | `/payments` | `X-User-ID`, `Idempotency-Key` | Create payment |
| GET | `/payments/:id` | `X-User-ID` | Get payment by ID |
| POST | `/refunds` | `X-User-ID` | Refund a successful payment |
| GET | `/health` | — | Health check |
| GET | `/metrics` | — | Prometheus metrics |

## Payment State Machine

```mermaid
stateDiagram-v2
    [*] --> CREATED
    CREATED --> AUTHORIZED: capture_mode_manual
    AUTHORIZED --> PROCESSING: capture
    CREATED --> PROCESSING: payment.created
    PROCESSING --> SUCCESS: provider.success
    PROCESSING --> FAILED: fraud_rejected_or_provider.failed
    SUCCESS --> CAPTURED: ledger_and_notification
    CAPTURED --> REFUNDED: refund_saga
    SUCCESS --> DISPUTED: dispute_opened
    FAILED --> [*]
    REFUNDED --> [*]
```

## New APIs (enhancements)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/payment-methods` | Tokenize payment method |
| GET | `/payments` | List user payments |
| POST | `/payments/:id/capture` | Capture authorized hold |
| POST | `/disputes` | Open dispute |

Refund saga enabled by default (`FEATURE_REFUND_SAGA=false` to disable).

## Saga Workflow

```mermaid
sequenceDiagram
    participant API as Payment_API
    participant DB as PostgreSQL
    participant Outbox as Outbox_Worker
    participant Kafka
    participant Fraud
    participant Provider
    participant Ledger
    participant Notify as Notification

    API->>DB: INSERT payment + outbox (txn)
    API-->>Client: 201 CREATED

    Outbox->>Kafka: payment.created
    Kafka->>API: Saga consumer
    API->>DB: status = PROCESSING
    API->>Kafka: fraud.validate

    Kafka->>Fraud: fraud.validate
    Fraud->>Kafka: fraud.validated

    Kafka->>API: fraud.validated
    alt approved
        API->>Kafka: provider.charge
        Kafka->>Provider: charge
        Provider->>Kafka: provider.success
        Kafka->>API: provider.success
        API->>DB: status = SUCCESS
        API->>Kafka: ledger.entry.created
        Kafka->>Ledger: write entry
        Kafka->>API: ledger.entry.created
        API->>Kafka: notification.send
        Kafka->>Notify: broadcast
    else rejected
        API->>DB: status = FAILED
        API->>Kafka: payment.failed
    end
```

```mermaid
flowchart TD
    CreatePayment[POST_payments] --> IdempotencyCheck{Redis_or_DB_idempotency}
    IdempotencyCheck -->|duplicate| ReturnCached[Return_cached_response]
    IdempotencyCheck -->|new| TxnWrite[DB_txn_payment_plus_outbox]
    TxnWrite --> OutboxWorker[Outbox_worker_poll]
    OutboxWorker --> PublishCreated[payment.created]
    PublishCreated --> FraudStep[fraud.validate]
    FraudStep --> ProviderStep[provider.charge]
    ProviderStep --> LedgerStep[ledger.entry.created]
    LedgerStep --> NotifyStep[notification.send]
```

## ERD

```mermaid
erDiagram
    payments ||--o{ refunds : has

    payments {
        uuid id PK
        uuid user_id
        numeric amount
        text currency
        text status
        text provider
        text idempotency_key UK
        timestamp created_at
        timestamp updated_at
    }

    refunds {
        uuid id PK
        uuid payment_id FK
        numeric amount
        text status
        timestamp created_at
    }

    outbox {
        uuid id PK
        uuid aggregate_id
        text event_type
        jsonb payload
        boolean processed
        timestamp created_at
    }
```

## Redis Keys

| Key | TTL | Purpose |
|-----|-----|---------|
| `idempotency:{key}` | 48h | Cached payment response |
| `idempotency:{key}:lock` | 30s | In-flight request lock |

## Kafka Topics

| Topic | Role |
|-------|------|
| `payment.created` | Produced (outbox), consumed (saga start) |
| `fraud.validate` | Produced |
| `fraud.validated` | Consumed |
| `provider.charge` | Produced |
| `provider.success` / `provider.failed` | Consumed |
| `ledger.entry.created` | Produced & consumed |
| `notification.send` | Produced |
| `payment.failed` | Produced on failure |

## Run

```bash
HTTP_PORT=8002 go run ./cmd/server
```

Migrations: `migrations/001_init.sql`
