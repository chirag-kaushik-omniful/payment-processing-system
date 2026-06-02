# Notification Service

Delivers real-time payment status updates via WebSocket. Consumes Kafka notification events and broadcasts to connected clients.

| Property | Value |
|----------|-------|
| Port | `8007` |
| Stack | Gin, WebSocket, Kafka |
| Persistence | In-memory hub (no database) |

## Responsibilities

- Consume `notification.send` events
- Broadcast payment status to WebSocket subscribers
- Maintain in-memory connection hub

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ws` | WebSocket upgrade for real-time updates |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Workflow

```mermaid
sequenceDiagram
    participant Payment as Payment_Saga
    participant Kafka
    participant Notify as Notification_Service
    participant Hub as WebSocket_Hub
    participant Client

    Client->>Notify: GET /ws (upgrade)
    Notify->>Hub: Register connection

    Payment->>Kafka: notification.send
    Kafka->>Notify: Consume event
    Notify->>Hub: Broadcast {payment_id, status, user_id}
    Hub->>Client: WebSocket message
```

```mermaid
flowchart TD
    ClientConnect[Client_CONNECT_ws] --> RegisterHub[Hub.Register]
    KafkaEvent[notification.send] --> ParsePayload[Parse_payment_id_status]
    ParsePayload --> Broadcast[Hub.Broadcast_all_clients]
    Broadcast --> WSMessage[JSON_over_WebSocket]
    ClientDisconnect[Client_DISCONNECT] --> UnregisterHub[Hub.Unregister]
```

## WebSocket Message Format

```json
{
  "payment_id": "uuid",
  "status": "completed",
  "user_id": "uuid"
}
```

## ERD

No persistent storage. In-memory connection model:

```mermaid
erDiagram
    WEBSOCKET_CLIENT {
        string connection_id PK
        string user_id
        timestamp connected_at
    }

    NOTIFICATION_EVENT {
        string payment_id
        string status
        string user_id
    }

    HUB {
        int active_connections
    }

    HUB ||--o{ WEBSOCKET_CLIENT : manages
    NOTIFICATION_EVENT ||--o{ WEBSOCKET_CLIENT : broadcast_to
```

## Kafka Topics

| Topic | Direction |
|-------|-----------|
| `notification.send` | Consume |

## Run

```bash
HTTP_PORT=8007 KAFKA_BROKERS=localhost:9092 go run ./cmd/server
```

Connect via: `ws://localhost:8007/ws`
