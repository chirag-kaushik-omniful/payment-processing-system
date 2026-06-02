# Auth Service

Handles user registration, login, JWT issuance, refresh token rotation, and role-based access control (RBAC).

| Property | Value |
|----------|-------|
| Port | `8001` |
| Stack | Gin, PostgreSQL, bcrypt, JWT |
| Persistence | PostgreSQL |

## Responsibilities

- User signup with bcrypt password hashing
- Login and JWT access/refresh token pair generation
- Refresh token rotation (revoke old, issue new)
- RBAC via `users`, `roles`, `user_roles`

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/signup` | Register new user |
| POST | `/login` | Authenticate and receive tokens |
| POST | `/refresh` | Rotate refresh token |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Workflow

```mermaid
sequenceDiagram
    participant Client
    participant Auth as Auth_Service
    participant DB as PostgreSQL

    Client->>Auth: POST /signup {email, password}
    Auth->>Auth: bcrypt hash password
    Auth->>DB: INSERT users + user_roles
    Auth->>Auth: Generate JWT pair
    Auth->>DB: INSERT refresh_tokens
    Auth-->>Client: access_token, refresh_token

    Client->>Auth: POST /login {email, password}
    Auth->>DB: SELECT user by email
    Auth->>Auth: bcrypt compare
    Auth->>DB: Revoke old refresh tokens
    Auth->>DB: INSERT new refresh_token
    Auth-->>Client: JWT pair
```

```mermaid
flowchart TD
    Signup[POST_signup] --> HashPassword[bcrypt_hash]
    HashPassword --> CreateUser[INSERT_users]
    CreateUser --> AssignRole[INSERT_user_roles]
    AssignRole --> IssueJWT[Generate_JWT_pair]
    IssueJWT --> StoreRefresh[INSERT_refresh_tokens]
    StoreRefresh --> Response[201_Created]
```

## ERD

```mermaid
erDiagram
    users ||--o{ user_roles : has
    roles ||--o{ user_roles : assigned
    users ||--o{ refresh_tokens : owns

    users {
        uuid id PK
        text email UK
        text password_hash
        timestamp created_at
    }

    roles {
        uuid id PK
        text name UK
    }

    user_roles {
        uuid user_id FK
        uuid role_id FK
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK
        text token_hash
        timestamp expires_at
        boolean revoked
        timestamp created_at
    }
```

## Default Roles

| ID | Name |
|----|------|
| `00000000-0000-0000-0000-000000000001` | `user` |
| `00000000-0000-0000-0000-000000000002` | `admin` |

## Run

```bash
HTTP_PORT=8001 DATABASE_URL=postgres://payment:payment@localhost:5432/payment_platform?sslmode=disable go run ./cmd/server
```

Migrations: `migrations/001_init.sql`
