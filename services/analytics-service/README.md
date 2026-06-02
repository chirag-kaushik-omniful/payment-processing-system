# Analytics Service

Aggregates payment metrics in memory from Kafka events and exposes a summary API for dashboards and monitoring.

| Property | Value |
|----------|-------|
| Port | `8010` |
| Stack | Gin, Kafka |
| Persistence | In-memory counters |

## Responsibilities

- Consume payment lifecycle events
- Track totals: payments created, completed, failed
- Track volume: total and completed amounts
- Expose aggregated summary via REST

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/analytics/summary` | Payment metrics summary |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Summary Response

```json
{
  "total_payments": 150,
  "completed_count": 120,
  "failed_count": 10,
  "total_volume": 45000.00,
  "completed_volume": 38000.00
}
```

## Workflow

```mermaid
sequenceDiagram
    participant Payment as Payment_Platform
    participant Kafka
    participant Analytics as Analytics_Service
    participant Dashboard

    Payment->>Kafka: payment.created
    Kafka->>Analytics: total_payments++, total_volume+=

    Payment->>Kafka: payment.completed
    Kafka->>Analytics: completed_count++, completed_volume+=

    Payment->>Kafka: payment.failed
    Kafka->>Analytics: failed_count++

    Dashboard->>Analytics: GET /analytics/summary
    Analytics-->>Dashboard: aggregated metrics
```

```mermaid
flowchart TD
    PaymentCreated[payment.created] --> IncTotal[total_payments++]
    IncTotal --> IncVolume[total_volume += amount]

    PaymentCompleted[payment.completed] --> IncCompleted[completed_count++]
    IncCompleted --> IncCompletedVol[completed_volume += amount]

    PaymentFailed[payment.failed] --> IncFailed[failed_count++]

    APISummary[GET_analytics_summary] --> ReadLock[RWMutex_read]
    ReadLock --> ReturnJSON[Return_PaymentSummary]
```

## ERD (In-Memory)

```mermaid
erDiagram
    PAYMENT_SUMMARY {
        int64 total_payments
        int64 completed_count
        int64 failed_count
        float total_volume
        float completed_volume
    }

    PAYMENT_CREATED_EVENT {
        string payment_id
        float amount
    }

    PAYMENT_COMPLETED_EVENT {
        string payment_id
        float amount
    }

    PAYMENT_FAILED_EVENT {
        string payment_id
    }

    PAYMENT_CREATED_EVENT ||--o{ PAYMENT_SUMMARY : increments
    PAYMENT_COMPLETED_EVENT ||--o{ PAYMENT_SUMMARY : increments
    PAYMENT_FAILED_EVENT ||--o{ PAYMENT_SUMMARY : increments
```

## Kafka Topics

| Topic | Counter updated |
|-------|-----------------|
| `payment.created` | `total_payments`, `total_volume` |
| `payment.completed` | `completed_count`, `completed_volume` |
| `payment.failed` | `failed_count` |

## Run

```bash
HTTP_PORT=8010 KAFKA_BROKERS=localhost:9092 go run ./cmd/server
```
