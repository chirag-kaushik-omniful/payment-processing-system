# Payment Platform — Complete Workflow

End-to-end flows across all microservices. For per-service detail, see each [service README](README.md#architecture).

---

## 1. System Overview

```mermaid
flowchart TB
    subgraph clients [Clients]
        WebApp[Web_App]
        MobileApp[Mobile_App]
        ProviderCallback[Provider_Webhooks]
    end

    subgraph edge [Edge]
        Gateway[API_Gateway_8000]
    end

    subgraph core [Core]
        Auth[Auth_8001]
        Payment[Payment_8002]
    end

    subgraph financial [Financial]
        Provider[Provider_8003]
        Wallet[Wallet_8004]
        Ledger[Ledger_8005]
    end

    subgraph risk_notify [Risk_and_Notify]
        Fraud[Fraud_8006]
        Notification[Notification_8007]
    end

    subgraph ops [Operations]
        Webhook[Webhook_8008]
        Reconciliation[Reconciliation_8009]
        Analytics[Analytics_8010]
    end

    subgraph infra [Infrastructure]
        Postgres[(PostgreSQL)]
        Redis[(Redis)]
        Kafka[(Kafka)]
    end

    WebApp --> Gateway
    MobileApp --> Gateway
    ProviderCallback --> Webhook

    Gateway --> Auth
    Gateway --> Payment
    Gateway --> Wallet

    Auth --> Postgres
    Payment --> Postgres
    Payment --> Redis
    Payment --> Kafka
    Wallet --> Postgres
    Wallet --> Redis
    Ledger --> Postgres

    Payment --> Kafka
    Fraud --> Kafka
    Fraud --> Redis
    Provider --> Kafka
    Ledger --> Kafka
    Notification --> Kafka
    Analytics --> Kafka
    Webhook --> Redis

    Kafka --> Fraud
    Kafka --> Provider
    Kafka --> Ledger
    Kafka --> Notification
    Kafka --> Analytics
    Kafka --> Wallet
```

---

## 2. Authentication Flow

Public routes on the gateway proxy to auth-service. Protected routes require a valid JWT.

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Gateway as API_Gateway
    participant Auth as Auth_Service
    participant DB as PostgreSQL

    Client->>Gateway: POST /auth/signup
    Gateway->>Auth: POST /signup
    Auth->>Auth: bcrypt hash password
    Auth->>DB: INSERT users, user_roles
    Auth->>Auth: Generate JWT pair
    Auth->>DB: INSERT refresh_tokens
    Auth-->>Gateway: 201 tokens
    Gateway-->>Client: access_token, refresh_token

    Client->>Gateway: POST /auth/login
    Gateway->>Auth: POST /login
    Auth->>DB: SELECT user, verify password
    Auth->>DB: Revoke old refresh tokens
    Auth->>DB: INSERT new refresh_token
    Auth-->>Gateway: JWT pair
    Gateway-->>Client: 200 OK

    Note over Client,Gateway: Subsequent requests
    Client->>Gateway: POST /payments (Authorization Bearer)
    Gateway->>Gateway: Validate JWT, rate limit
    Gateway->>Gateway: Set X-User-ID header
```

---

## 3. Create Payment (HTTP + Outbox)

Synchronous API write uses the **transactional outbox** so Kafka publish is never lost.

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Gateway as API_Gateway
    participant Payment as Payment_Service
    participant Redis
    participant DB as PostgreSQL
    participant Outbox as Outbox_Worker
    participant Kafka

    Client->>Gateway: POST /payments + Idempotency-Key
    Gateway->>Gateway: JWT + rate limit
    Gateway->>Payment: POST /payments (X-User-ID)

    Payment->>Redis: GET idempotency:key
    alt duplicate request
        Redis-->>Payment: cached response
        Payment-->>Client: 201 cached
    else new payment
        Payment->>DB: BEGIN TX
        Payment->>DB: INSERT payments (CREATED)
        Payment->>DB: INSERT outbox (payment.created)
        Payment->>DB: COMMIT
        Payment->>Redis: SET idempotency:key
        Payment-->>Client: 201 payment_id
    end

    loop every 2s
        Outbox->>DB: SELECT unprocessed outbox
        Outbox->>Kafka: Publish payment.created
        Outbox->>DB: Mark outbox processed
    end
```

---

## 4. Payment Saga — Success Path

Orchestrated by **payment-service** via Kafka. Each step is async and idempotent.

```mermaid
sequenceDiagram
    autonumber
    participant Kafka
    participant Payment as Payment_Service
    participant Fraud as Fraud_Service
    participant Provider as Provider_Service
    participant Ledger as Ledger_Service
    participant Notification as Notification_Service
    participant Analytics as Analytics_Service
    participant DB as PostgreSQL

    Kafka->>Payment: payment.created
    Payment->>DB: status = PROCESSING
    Payment->>Kafka: fraud.validate

    Kafka->>Fraud: fraud.validate
    Fraud->>Fraud: velocity + amount rules
    Fraud->>Kafka: fraud.validated (approved=true)

    Kafka->>Payment: fraud.validated
    Payment->>Kafka: provider.charge

    Kafka->>Provider: provider.charge
    Provider->>Provider: Stripe/Razorpay Charge
    Provider->>Kafka: provider.success

    Kafka->>Payment: provider.success
    Payment->>DB: status = SUCCESS
    Payment->>Kafka: ledger.entry.created

    Kafka->>Ledger: ledger.entry.created
    Ledger->>DB: INSERT ledger_entries (append-only)

    Kafka->>Payment: ledger.entry.created
    Payment->>Kafka: notification.send

    Kafka->>Notification: notification.send
    Notification->>Notification: WebSocket broadcast

    Note over Analytics: Parallel consumers
    Kafka->>Analytics: payment.created / completed
```

```mermaid
flowchart LR
    A[payment.created] --> B[fraud.validate]
    B --> C[fraud.validated]
    C --> D[provider.charge]
    D --> E[provider.success]
    E --> F[ledger.entry.created]
    F --> G[notification.send]
```

---

## 5. Payment Saga — Failure Paths

```mermaid
flowchart TD
    Start[payment.created] --> FraudCheck[fraud.validate]

    FraudCheck --> FraudResult{fraud.validated}

    FraudResult -->|approved=false| FailFraud[status FAILED]
    FailFraud --> PubFailed1[payment.failed]

    FraudResult -->|approved=true| ProviderCharge[provider.charge]
    ProviderCharge --> ProviderResult{provider result}

    ProviderResult -->|provider.failed| FailProvider[status FAILED]
    FailProvider --> PubFailed2[payment.failed]

    ProviderResult -->|provider.success| Success[status SUCCESS]
    Success --> Ledger[ledger.entry.created]
    Ledger --> Notify[notification.send]

    PubFailed1 --> AnalyticsFail[analytics: failed_count++]
    PubFailed2 --> AnalyticsFail
```

```mermaid
sequenceDiagram
    autonumber
    participant Kafka
    participant Payment as Payment_Service
    participant Fraud as Fraud_Service
    participant Provider as Provider_Service
    participant Analytics as Analytics_Service

    alt fraud rejected
        Kafka->>Fraud: fraud.validate
        Fraud->>Kafka: fraud.validated approved=false
        Kafka->>Payment: fraud.validated
        Payment->>Payment: status = FAILED
        Payment->>Kafka: payment.failed
        Kafka->>Analytics: payment.failed
    else provider charge failed
        Kafka->>Provider: provider.charge
        Provider->>Kafka: provider.failed
        Kafka->>Payment: provider.failed
        Payment->>Payment: status = FAILED
        Payment->>Kafka: payment.failed
        Kafka->>Analytics: payment.failed
    end
```

---

## 6. Payment State Machine

```mermaid
stateDiagram-v2
    [*] --> CREATED: POST /payments

    CREATED --> PROCESSING: payment.created consumed

    PROCESSING --> SUCCESS: provider.success
    PROCESSING --> FAILED: fraud rejected OR provider.failed

    SUCCESS --> REFUNDED: POST /refunds

    FAILED --> [*]
    REFUNDED --> [*]

    note right of CREATED
        Outbox publishes payment.created
    end note

    note right of SUCCESS
        Ledger + notification follow
    end note
```

| Status | Trigger |
|--------|---------|
| `CREATED` | Payment inserted in DB |
| `PROCESSING` | Saga consumes `payment.created` |
| `SUCCESS` | `provider.success` received |
| `FAILED` | Fraud rejected or `provider.failed` |
| `REFUNDED` | `POST /refunds` on successful payment |

---

## 7. Refund Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Gateway as API_Gateway
    participant Payment as Payment_Service
    participant DB as PostgreSQL

    Client->>Gateway: POST /refunds (JWT)
    Gateway->>Payment: POST /refunds (X-User-ID)
    Payment->>DB: SELECT payment (must be SUCCESS)
    Payment->>DB: INSERT refunds
    Payment->>DB: UPDATE payments status = REFUNDED
    Payment-->>Client: 200 payment response
```

---

## 8. Wallet Flow

Wallet updates happen asynchronously via Kafka (and synchronously via `GET /wallet`).

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Gateway as API_Gateway
    participant Wallet as Wallet_Service
    participant Redis
    participant DB as PostgreSQL
    participant Kafka

    Client->>Gateway: GET /wallet (JWT)
    Gateway->>Wallet: GET /wallet
    Wallet->>Redis: GET wallet:balance:user_id
    alt cache miss
        Wallet->>DB: SELECT wallets
        Wallet->>Redis: SET cache
    end
    Wallet-->>Client: balance, version

    Kafka->>Wallet: wallet.credited
    Wallet->>DB: UPDATE balance (optimistic lock)
    Wallet->>Redis: invalidate cache

    Kafka->>Wallet: wallet.debited
    Wallet->>DB: UPDATE balance (optimistic lock)
```

---

## 9. Provider Webhook Flow

External providers call webhook-service directly (not through the gateway).

```mermaid
sequenceDiagram
    autonumber
    participant Provider as Stripe_Razorpay
    participant Webhook as Webhook_Service
    participant Redis

    Provider->>Webhook: POST /webhooks/stripe + signature
    Webhook->>Webhook: HMAC-SHA256 verify
    alt invalid signature
        Webhook-->>Provider: 401
    else valid
        Webhook->>Redis: idempotency lookup
        alt already processed
            Webhook-->>Provider: 200 replay=true
        else new event
            Webhook->>Redis: lock + store response
            Webhook-->>Provider: 200 processed=true
        end
    end
```

---

## 10. Reconciliation Flow

Operations team compares provider settlement files against ledger exports.

```mermaid
sequenceDiagram
    autonumber
    participant Ops as Operations
    participant Provider as Provider_Settlement
    participant Ledger as Ledger_Service
    participant Recon as Reconciliation_Service

    Provider->>Ops: Settlement file
    Ops->>Ledger: Export ledger entries
    Ops->>Recon: POST /reconcile {settlements, ledger}
    Recon->>Recon: Match by payment_id
    Recon->>Recon: Detect missing / amount mismatch
    Recon-->>Ops: matched, mismatches[]
```

---

## 11. Analytics Flow

Analytics-service listens in parallel — it does not block the saga.

```mermaid
flowchart LR
    payment_created[payment.created] --> Analytics[Analytics_8010]
    payment_completed[payment.completed] --> Analytics
    payment_failed[payment.failed] --> Analytics
    Analytics --> API[GET /analytics/summary]
```

| Event | Metrics updated |
|-------|-----------------|
| `payment.created` | `total_payments++`, `total_volume += amount` |
| `payment.completed` | `completed_count++`, `completed_volume += amount` |
| `payment.failed` | `failed_count++` |

---

## 12. Kafka Topic Map

All topics partition by `payment_id` for ordering.

```mermaid
flowchart TB
    subgraph payment_topics [Payment_Lifecycle]
        T1[payment.created]
        T2[payment.processing]
        T3[payment.completed]
        T4[payment.failed]
        T5[refund.requested]
    end

    subgraph saga_topics [Saga_Orchestration]
        S1[fraud.validate]
        S2[fraud.validated]
        S3[provider.charge]
        S4[provider.success]
        S5[provider.failed]
        S6[ledger.entry.created]
        S7[notification.send]
    end

    subgraph wallet_topics [Wallet]
        W1[wallet.credited]
        W2[wallet.debited]
    end

    subgraph ops_topics [Operations]
        O1[reconciliation.started]
    end

    T1 --> S1
    S1 --> S2
    S2 --> S3
    S3 --> S4
    S3 --> S5
    S4 --> S6
    S6 --> S7
```

| Topic | Producer | Consumer(s) |
|-------|----------|-------------|
| `payment.created` | Outbox worker | payment-service (saga), analytics-service |
| `fraud.validate` | payment-service | fraud-service |
| `fraud.validated` | fraud-service | payment-service |
| `provider.charge` | payment-service | provider-service |
| `provider.success` | provider-service | payment-service |
| `provider.failed` | provider-service | payment-service |
| `ledger.entry.created` | payment-service | ledger-service, payment-service |
| `notification.send` | payment-service | notification-service |
| `payment.failed` | payment-service | analytics-service |
| `wallet.credited` | (external) | wallet-service |
| `wallet.debited` | (external) | wallet-service |

Retry/DLQ pattern: `{topic}.retry` → `{topic}.dlq` (see `shared/kafka`).

---

## 13. End-to-End Happy Path (Single Diagram)

```mermaid
sequenceDiagram
    autonumber
    box Client
        participant User
    end
    box Edge_8000
        participant Gateway
    end
    box Services
        participant Auth
        participant Payment
        participant Fraud
        participant Provider
        participant Ledger
        participant Notify as Notification
        participant Analytics
    end
    box Data
        participant DB
        participant Redis
        participant Kafka
    end

    User->>Gateway: POST /auth/login
    Gateway->>Auth: login
    Auth->>DB: verify user
    Auth-->>User: JWT

    User->>Gateway: POST /payments
    Gateway->>Payment: create payment
    Payment->>Redis: idempotency check
    Payment->>DB: payment + outbox
    Payment-->>User: 201 CREATED

    Payment->>Kafka: payment.created
    Payment->>Kafka: fraud.validate
    Kafka->>Fraud: validate
    Fraud->>Kafka: fraud.validated
    Payment->>Kafka: provider.charge
    Kafka->>Provider: charge
    Provider->>Kafka: provider.success
    Payment->>DB: SUCCESS
    Payment->>Kafka: ledger.entry.created
    Kafka->>Ledger: double-entry write
    Payment->>Kafka: notification.send
    Kafka->>Notify: WebSocket push
    Kafka->>Analytics: update counters
    Notify-->>User: payment completed
```

---

## 14. Service Ports Quick Reference

| Port | Service | Role in workflow |
|------|---------|------------------|
| 8000 | api-gateway | Entry, JWT, rate limit, proxy |
| 8001 | auth-service | Identity and tokens |
| 8002 | payment-service | Orchestrator, outbox, saga |
| 8003 | provider-service | External charge |
| 8004 | wallet-service | Balance read/update |
| 8005 | ledger-service | Accounting entries |
| 8006 | fraud-service | Risk gate |
| 8007 | notification-service | Real-time alerts |
| 8008 | webhook-service | Provider callbacks |
| 8009 | reconciliation-service | Settlement audit |
| 8010 | analytics-service | Metrics aggregation |

---

## 15. Refund Saga

```mermaid
sequenceDiagram
    participant Client
    participant Payment
    participant Kafka
    participant Provider
    participant Ledger

    Client->>Payment: POST /refunds (admin)
    Payment->>Kafka: refund.requested
    Kafka->>Provider: provider.refund
    Provider->>Kafka: refund.completed
    Kafka->>Payment: update REFUNDED
    Payment->>Kafka: ledger.compensation
    Kafka->>Ledger: append compensation entry
```

## 16. Auto Reconciliation

```mermaid
sequenceDiagram
    participant Ops
    participant Recon as Reconciliation_8009
    participant Ledger as Ledger_8005

    Ops->>Recon: POST /reconcile/auto {settlements}
    Recon->>Ledger: GET /ledger/export
    Ledger-->>Recon: ledger rows
    Recon->>Recon: compare by payment_id
    Recon-->>Ops: matched / mismatches
```

## 17. Webhook to Saga

```mermaid
sequenceDiagram
    participant Provider
    participant Webhook as Webhook_8008
    participant Kafka
    participant Payment

    Provider->>Webhook: POST /webhooks/stripe
    Webhook->>Webhook: HMAC verify + idempotency
    Webhook->>Kafka: webhook.received
    Kafka->>Payment: update status / dispute
```

## Related Docs

- [README.md](README.md) — Quick start and project layout
- [docs/ENHANCEMENTS.md](docs/ENHANCEMENTS.md) — All implemented enhancements
- [docs/architecture.md](docs/architecture.md) — Local development setup
- [.md](.md) — Full platform blueprint (section 19: enhancements)
- Per-service READMEs under `services/*/README.md`
