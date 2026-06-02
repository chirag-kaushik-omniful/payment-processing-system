# Subscription Service

Recurring billing: subscription CRUD and scheduled `subscription.charge` Kafka events.

| Property | Value |
|----------|-------|
| Port | `8012` |
| Persistence | PostgreSQL |

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/subscriptions` | Create subscription |
| GET | `/subscriptions/:id` | Get subscription |

## ERD

```mermaid
erDiagram
    subscriptions {
        uuid id PK
        uuid user_id
        text plan_id
        numeric amount
        text currency
        text interval
        text status
        timestamp next_billing_at
    }
```

## Workflow

```mermaid
flowchart LR
    Scheduler[Minute_ticker] --> Due[Find_due_subscriptions]
    Due --> Kafka[subscription.charge]
    Kafka --> Payment[Payment_processing]
```

Migrations: `migrations/001_subscriptions.sql`
