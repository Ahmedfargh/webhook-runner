# Webhook Runner Service

High-throughput gRPC microservice built with Go, GORM, Protobuf, and HMVC architecture for managing user applications, HMAC-SHA256 cryptographic signatures, reliable HTTP webhook dispatching, retry lifecycles, and execution logs telemetry.

---

## 🏛 HMVC Architecture

The codebase follows the Hierarchical Model-View-Controller (HMVC) design pattern:

```
webhook-runner/
├── api/
│   └── proto/v1/          # Protobuf definitions & generated gRPC code
├── cmd/
│   ├── server/            # Microservice gRPC entrypoint (Port 50053)
│   ├── migrations/        # GORM database schema auto-migrations
│   └── seeders/           # Realistic developer applications & webhook seeders
└── internal/
    ├── config/            # Environment parsing & database connection pooling
    ├── engine/            # HTTP Dispatcher & HMAC-SHA256 crypto signing engine
    ├── middleware/        # gRPC caller service whitelist & Bearer token interceptor
    ├── models/            # GORM entities (App, WebhookCall)
    └── modules/
        ├── app/           # App registration, credentials, secret rotation
        └── webhook/       # Webhook dispatching, telemetry logs, retry engine
```

---

## 🔐 Cryptographic Signing & Headers

Every outgoing HTTP webhook request is automatically signed using the app's `webhook_secret` with **HMAC-SHA256**:

### Injected Request Headers:
| Header | Description | Example |
| :--- | :--- | :--- |
| `X-Webhook-Signature` | Hex-encoded HMAC-SHA256 signature | `sha256=9b7c8...` |
| `X-Webhook-ID` | Unique execution call identifier | `wh_call_88291...` |
| `X-Webhook-Event` | Event topic name | `order.created` |
| `X-Webhook-Timestamp` | Unix epoch timestamp (seconds) | `1788645000` |
| `X-Request-ID` | Distributed end-to-end request correlation ID | `req-018e4b7a...` |
| `Content-Type` | JSON payload MIME type | `application/json` |

---

## ⚡ Asynchronous Kafka Ingestion & Trace Propagation

In addition to direct gRPC dispatching, `webhook-runner` consumes events from Apache Kafka:
- **Topic `webhook-dispatches`**: High-throughput message queue for background dispatching.
- **Trace Propagation**: Inherits `request_id` / `trace_id` from Kafka message headers and carries it into outgoing HTTP webhook requests (`X-Request-ID`), ensuring end-to-end APM visibility from caller to destination.
- **Audit Emission**: Asynchronously emits application mutations and secret rotations to Kafka topic `audit-events`.

---

## 📊 Database Schema (`webhook_runner`)

### 1. `apps`
- `id`: UUID (Primary Key)
- `user_id`: UUID (Owner account reference)
- `name`: Application Name
- `app_id`: Unique client identifier string (e.g. `app_live_f4735dd...`)
- `app_secret`: Cryptographic client secret
- `webhook_url`: Default HTTP/HTTPS destination URL
- `webhook_secret`: Secret key used for HMAC-SHA256 request signing
- `is_active`: Boolean status flag
- `created_at`, `updated_at`, `deleted_at`

### 2. `webhook_calls`
- `id`: UUID (Primary Key)
- `app_id`: UUID (Foreign Key $\rightarrow$ `apps.id`)
- `event_name`: Event name (e.g. `order.created`)
- `target_url`: Target HTTP URL where the webhook was dispatched
- `payload_json`: LongText raw JSON payload data
- `headers_json`: Custom request headers
- `status`: Execution status (`SUCCESS`, `FAILED`, `PENDING`, `RETRYING`, `TIMEOUT`)
- `response_status_code`: Downstream HTTP response code (e.g. `200`, `500`)
- `response_body`: Downstream HTTP response body
- `response_latency_ms`: Round-trip execution time in milliseconds
- `error_message`: Network / DNS error message (if delivery failed)
- `attempt_count`: Total delivery attempts
- `next_retry_at`: Next retry scheduled timestamp
- `created_at`, `updated_at`

---

## 🚀 Running the Service

### 1. Run Migrations & Seeders
```bash
# Run database migrations
go run cmd/migrations/migrate.go

# Seed realistic developer apps and mock deliveries
go run cmd/seeders/Seeder.go
```

### 2. Start the gRPC Server
```bash
go run cmd/server/main.go
```
The server will start on port `50053` with gRPC Reflection enabled.

---

## 📡 Key gRPC Methods

- `SendWebhook`: Dispatch a webhook event with HMAC-SHA256 signature and record telemetry.
- `ListWebhookCalls`: Retrieve execution logs filtered by `app_id`, `status`, `search`, and pagination.
- `GetWebhookCall`: Inspect full payload, headers, response code, and latency for a specific call.
- `RetryWebhookCall`: Re-dispatch a previously failed webhook attempt.
- `CreateApp`, `GetApp`, `ListApps`, `UpdateApp`, `DeleteApp`: Application credentials lifecycle.
- `RotateSecrets`: Re-generate client secret and/or HMAC secret.
