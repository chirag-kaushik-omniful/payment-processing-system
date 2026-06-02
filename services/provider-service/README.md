# Provider Service

Integrates with external payment providers via the **Factory**, **Strategy**, and **Adapter** patterns. Ships with mock Stripe and Razorpay implementations for local development.

| Property | Value |
|----------|-------|
| Port | `8003` |
| Stack | Kafka, Factory/Strategy pattern |
| Persistence | None (stateless) |

## Responsibilities

- Consume `provider.charge` events from the payment saga
- Route charges to the correct provider adapter (Stripe, Razorpay)
- Publish `provider.success` or `provider.failed`
- Record provider latency metrics

## Supported Providers

| Provider | Package | Behavior |
|----------|---------|----------|
| `stripe` | `internal/provider/stripe` | Mock charge (always succeeds) |
| `razorpay` | `internal/provider/razorpay` | Mock charge (always succeeds) |

## Provider Interface

```go
type PaymentProvider interface {
    Name() string
    Charge(ctx context.Context, req ChargeRequest) (transactionID string, err error)
}
```

## Workflow

```mermaid
sequenceDiagram
    participant Kafka
    participant Provider as Provider_Service
    participant Factory
    participant Stripe
    participant Payment as Payment_Saga

    Kafka->>Provider: provider.charge
    Provider->>Factory: Get(provider_name)
    Factory->>Stripe: Charge(amount, currency)
    Stripe-->>Provider: transaction_id
    Provider->>Provider: Record latency metric
    Provider->>Kafka: provider.success
    Kafka->>Payment: Continue saga
```

```mermaid
flowchart TD
    ConsumeCharge[Consume_provider.charge] --> ParseEvent[Parse_payment_id_amount]
    ParseEvent --> FactorySelect[Factory.Get_provider]
    FactorySelect --> StripeAdapter[Stripe_Adapter]
    FactorySelect --> RazorpayAdapter[Razorpay_Adapter]
    StripeAdapter --> ChargeCall[Charge_API_mock]
    RazorpayAdapter --> ChargeCall
    ChargeCall -->|success| PublishSuccess[provider.success]
    ChargeCall -->|failure| PublishFailed[provider.failed]
```

## ERD

No relational database. Conceptual model of provider charge events:

```mermaid
erDiagram
    CHARGE_EVENT {
        string payment_id PK
        string provider
        float amount
        string currency
    }

    CHARGE_RESULT {
        string payment_id PK
        string provider
        string transaction_id
        boolean success
        string failure_reason
    }

    CHARGE_EVENT ||--o| CHARGE_RESULT : produces
```

## Kafka Topics

| Topic | Direction |
|-------|-----------|
| `provider.charge` | Consume |
| `provider.success` | Produce |
| `provider.failed` | Produce |

## Metrics

- `provider_latency_seconds{provider="stripe|razorpay"}`

## Run

```bash
HTTP_PORT=8003 KAFKA_BROKERS=localhost:9092 go run ./cmd/server
```
