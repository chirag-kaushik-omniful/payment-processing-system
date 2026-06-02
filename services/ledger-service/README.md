# Ledger Service

Immutable, append-only **double-entry accounting** ledger. Every financial movement is recorded as balanced debit/credit entries and never updated in place.

| Property | Value |
|----------|-------|
| Port | `8005` |
| Stack | PostgreSQL, Kafka |
| Persistence | PostgreSQL |

## Responsibilities

- Record double-entry ledger rows on `ledger.entry.created` events
- Enforce append-only writes (no UPDATE/DELETE on entries)
- Idempotent processing by `transaction_id` (payment_id)
- Default accounts: `wallet:{user_id}` → `revenue:payments`

## Workflow

```mermaid
sequenceDiagram
    participant Payment as Payment_Saga
    participant Kafka
    participant Ledger as Ledger_Service
    participant DB as PostgreSQL

    Payment->>Kafka: ledger.entry.created
    Kafka->>Ledger: Consume event
    Ledger->>DB: EXISTS transaction_id?
    alt already recorded
        Ledger-->>Kafka: skip (idempotent)
    else new
        Ledger->>DB: INSERT ledger_entry
        Note over DB: debit=wallet:user_id<br/>credit=revenue:payments
    end
```

```mermaid
flowchart TD
    ConsumeEvent[Consume_ledger.entry.created] --> IdempotencyCheck{Entry_exists_for_payment_id}
    IdempotencyCheck -->|yes| Skip[Return_idempotent]
    IdempotencyCheck -->|no| BuildEntry[Build_debit_credit_accounts]
    BuildEntry --> AppendOnly[INSERT_ledger_entries]
    AppendOnly --> Done[Ack_message]
```

## Double-Entry Model

Every payment produces one balanced entry:

| Debit Account | Credit Account | Meaning |
|---------------|----------------|---------|
| `wallet:{user_id}` | `revenue:payments` | Funds moved from user wallet to revenue |

```mermaid
flowchart LR
    DebitAccount["wallet:user_id"] -->|amount| CreditAccount["revenue:payments"]
```

## ERD

```mermaid
erDiagram
    ledger_entries {
        uuid id PK
        uuid transaction_id "payment_id, idempotency key"
        text debit_account
        text credit_account
        numeric amount
        timestamp created_at
    }
```

| Constraint | Rule |
|------------|------|
| Append-only | No UPDATE or DELETE on `ledger_entries` |
| Balance | Each row represents one side of a balanced transaction |
| Idempotency | Unique processing per `transaction_id` |
| Amount | `CHECK (amount > 0)` |

## Indexes

- `idx_ledger_entries_transaction_id` — idempotency lookups
- `idx_ledger_entries_created_at` — audit/time-range queries

## Kafka Topics

| Topic | Direction |
|-------|-----------|
| `ledger.entry.created` | Consume |

## Run

```bash
HTTP_PORT=8005 go run ./cmd/server
```

Migrations: `migrations/001_init.sql`
