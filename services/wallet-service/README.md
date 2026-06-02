# Wallet Service

Manages user wallet balances with **optimistic locking** for concurrent updates and Redis caching for fast reads.

| Property | Value |
|----------|-------|
| Port | `8004` |
| Stack | Gin, PostgreSQL, Redis, Kafka |
| Persistence | PostgreSQL + Redis |

## Responsibilities

- Wallet balance reads (`GET /wallet`)
- Credit and debit operations with version-based optimistic locking
- Redis cache for balance (`wallet:balance:{user_id}`)
- Kafka consumers for `wallet.credited` and `wallet.debited` events

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/wallet` | Get balance (header `X-User-ID` or `?user_id=`) |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Workflow

```mermaid
sequenceDiagram
    participant Client
    participant Wallet as Wallet_Service
    participant Redis
    participant DB as PostgreSQL

    Client->>Wallet: GET /wallet
    Wallet->>Redis: GET wallet:balance:{user_id}
    alt cache hit
        Redis-->>Wallet: balance
    else cache miss
        Wallet->>DB: SELECT wallets
        DB-->>Wallet: balance, version
        Wallet->>Redis: SET cache
    end
    Wallet-->>Client: {balance, version}
```

```mermaid
sequenceDiagram
    participant Kafka
    participant Wallet as Wallet_Service
    participant DB as PostgreSQL

    Kafka->>Wallet: wallet.credited {user_id, amount}
    loop optimistic retry up to 5
        Wallet->>DB: SELECT balance, version
        Wallet->>DB: UPDATE WHERE version = N
        alt version conflict
            Wallet->>Wallet: retry
        else success
            Wallet->>Redis: invalidate cache
        end
    end
```

```mermaid
flowchart TD
    GetWallet[GET_wallet] --> CacheCheck{Redis_cache}
    CacheCheck -->|hit| ReturnBalance[Return_balance]
    CacheCheck -->|miss| DBRead[SELECT_from_wallets]
    DBRead --> PopulateCache[SET_wallet_balance_user_id]
    PopulateCache --> ReturnBalance

    KafkaCredit[wallet.credited] --> OptimisticUpdate[UPDATE_with_version_check]
    OptimisticUpdate -->|conflict| Retry[Retry_up_to_5]
    Retry --> OptimisticUpdate
    OptimisticUpdate -->|success| InvalidateCache[Invalidate_Redis]
```

## ERD

```mermaid
erDiagram
    wallets {
        uuid user_id PK
        numeric balance
        int version
        timestamp updated_at
    }
```

| Column | Description |
|--------|-------------|
| `version` | Incremented on every update; used for optimistic locking |
| `balance` | Current wallet balance (NUMERIC 18,2) |

## Optimistic Locking SQL

```sql
UPDATE wallets
SET balance = $1, version = version + 1, updated_at = NOW()
WHERE user_id = $2 AND version = $3;
```

## Redis Keys

| Key | Purpose |
|-----|---------|
| `wallet:balance:{user_id}` | Cached balance for reads |

## Kafka Topics

| Topic | Action |
|-------|--------|
| `wallet.credited` | Increase balance |
| `wallet.debited` | Decrease balance |

## Run

```bash
HTTP_PORT=8004 go run ./cmd/server
```

Migrations: `migrations/001_init.sql`
