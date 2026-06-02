# Reconciliation Service

Compares provider settlement records against ledger entries to detect mismatches and support recovery workflows.

| Property | Value |
|----------|-------|
| Port | `8009` |
| Stack | Gin |
| Persistence | None (stateless comparison) |

## Responsibilities

- Accept settlement file data and ledger snapshot via API
- Match payments by `payment_id`
- Detect missing entries and amount mismatches
- Return reconciliation report with mismatch details

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/reconcile` | Compare settlements vs ledger |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Request / Response

**POST /reconcile**

```json
{
  "settlements": [
    { "payment_id": "uuid", "amount": 100.00 }
  ],
  "ledger": [
    { "payment_id": "uuid", "amount": 100.00 }
  ]
}
```

```json
{
  "matched": 8,
  "mismatched": 2,
  "mismatches": [
    {
      "payment_id": "uuid",
      "settlement_amount": 100.00,
      "ledger_amount": 99.00,
      "reason": "amount mismatch"
    }
  ]
}
```

## Workflow

```mermaid
sequenceDiagram
    participant Ops as Operations_Team
    participant Recon as Reconciliation_Service
    participant Ledger as Ledger_Service
    participant Provider as Provider_Settlement

    Provider->>Ops: Settlement file (CSV/API)
    Ops->>Ledger: Export ledger entries
    Ops->>Recon: POST /reconcile {settlements, ledger}
    Recon->>Recon: Build payment_id maps
    Recon->>Recon: Compare amounts both directions
    Recon-->>Ops: matched + mismatches report
```

```mermaid
flowchart TD
    Start[POST_reconcile] --> IndexLedger[Map_ledger_by_payment_id]
    IndexLedger --> LoopSettlements[For_each_settlement]
    LoopSettlements --> InLedger{Exists_in_ledger}
    InLedger -->|no| MismatchMissing[reason_missing_in_ledger]
    InLedger -->|yes| AmountMatch{amounts_equal}
    AmountMatch -->|no| MismatchAmount[reason_amount_mismatch]
    AmountMatch -->|yes| IncrementMatched[matched++]
    LoopSettlements --> CheckOrphans[Ledger_entries_not_in_settlement]
    CheckOrphans --> MismatchOrphan[reason_missing_in_settlement]
    MismatchMissing --> Report[Return_mismatches]
    MismatchAmount --> Report
    MismatchOrphan --> Report
    IncrementMatched --> Report
```

## ERD (Logical)

Stateless service — input datasets are compared in memory:

```mermaid
erDiagram
    SETTLEMENT_RECORD {
        string payment_id PK
        float amount
        date settlement_date
    }

    LEDGER_RECORD {
        string payment_id PK
        float amount
    }

    MISMATCH {
        string payment_id
        float settlement_amount
        float ledger_amount
        string reason
    }

    RECONCILIATION_REPORT {
        int matched
        int mismatched
    }

    SETTLEMENT_RECORD ||--o| LEDGER_RECORD : compared_with
    SETTLEMENT_RECORD ||--o{ MISMATCH : may_produce
    LEDGER_RECORD ||--o{ MISMATCH : may_produce
    RECONCILIATION_REPORT ||--|{ MISMATCH : contains
```

## Mismatch Reasons

| Reason | Description |
|--------|-------------|
| `missing in ledger` | Settlement exists, no ledger entry |
| `amount mismatch` | Both exist, amounts differ |
| `missing in settlement` | Ledger entry with no settlement |

## Run

```bash
HTTP_PORT=8009 go run ./cmd/server
```
