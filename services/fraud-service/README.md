# Fraud Service

Redis-backed fraud detection with velocity rules and amount thresholds. Validates payments in the saga before provider charging.

| Property | Value |
|----------|-------|
| Port | `8006` |
| Stack | Redis, Kafka |
| Persistence | Redis only |

## Responsibilities

- Consume `fraud.validate` events from the payment saga
- Velocity checks per user and IP
- High-amount threshold rejection
- Publish `fraud.validated` with approval decision

## Risk Rules

| Rule | Threshold | Action |
|------|-----------|--------|
| User velocity | > 10 requests / minute | Reject |
| IP velocity | > 50 requests / minute | Reject |
| High amount | >= $10,000 | Reject |

## Workflow

```mermaid
sequenceDiagram
    participant Payment as Payment_Saga
    participant Kafka
    participant Fraud as Fraud_Service
    participant Redis

    Payment->>Kafka: fraud.validate
    Kafka->>Fraud: Consume event
    Fraud->>Fraud: Check amount threshold
    Fraud->>Redis: INCR fraud:user:{id}:velocity
    Fraud->>Redis: INCR fraud:ip:{ip}
    alt rules pass
        Fraud->>Kafka: fraud.validated {approved: true}
    else rule failed
        Fraud->>Kafka: fraud.validated {approved: false, reason}
    end
    Kafka->>Payment: Continue or fail saga
```

```mermaid
flowchart TD
    ConsumeValidate[Consume_fraud.validate] --> AmountCheck{amount_ge_10000}
    AmountCheck -->|yes| Reject[approved_false_high_amount]
    AmountCheck -->|no| UserVelocity[Redis_INCR_user_velocity]
    UserVelocity --> UserLimit{count_gt_10_per_min}
    UserLimit -->|yes| RejectVelocity[approved_false_velocity]
    UserLimit -->|no| IPVelocity[Redis_INCR_ip_velocity]
    IPVelocity --> IPLimit{count_gt_50_per_min}
    IPLimit -->|yes| RejectVelocity
    IPLimit -->|no| Approve[approved_true]
    Reject --> PublishValidated[fraud.validated]
    RejectVelocity --> PublishValidated
    Approve --> PublishValidated
```

## ERD (Redis)

No relational database. Velocity counters stored in Redis:

```mermaid
erDiagram
    USER_VELOCITY {
        string key PK "fraud:user:{user_id}:velocity"
        int count
        datetime ttl "1 minute"
    }

    IP_VELOCITY {
        string key PK "fraud:ip:{ip}"
        int count
        datetime ttl "1 minute"
    }

    FRAUD_VALIDATE_EVENT {
        string payment_id
        string user_id
        float amount
        string ip
    }

    FRAUD_VALIDATED_EVENT {
        string payment_id
        float amount
        boolean approved
        string reason
    }

    FRAUD_VALIDATE_EVENT ||--o| FRAUD_VALIDATED_EVENT : evaluates_to
```

## Redis Keys

| Key | TTL | Limit |
|-----|-----|-------|
| `fraud:user:{user_id}:velocity` | 1m | 10 |
| `fraud:ip:{ip}` | 1m | 50 |

## Kafka Topics

| Topic | Direction |
|-------|-----------|
| `fraud.validate` | Consume |
| `fraud.validated` | Produce |

## Run

```bash
HTTP_PORT=8006 REDIS_URL=redis://localhost:6379/0 KAFKA_BROKERS=localhost:9092 go run ./cmd/server
```
