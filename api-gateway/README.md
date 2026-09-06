# Webhook API Gateway

High-performance REST API Gateway written in Go (Gin) for orchestrating microservices, handling client authentication (JWT), terminating REST requests, and securely routing them to downstream microservices (`accounts`, `subscriptions`, `webhook-runner`) via gRPC.

---

## 🏛 Architecture & Security Features

- **Unified Public REST API**: Exposes clean JSON REST endpoints for Apps, Webhooks, Billing, Subscriptions, Invoices, Users, Admins, Roles, Permissions, Request Traces (APM), and Audit Logs.
- **End-to-End Request Tracing (APM)**: `RequestTrackerMiddleware` generates or extracts a unique `X-Request-ID` / `X-Trace-ID` for every HTTP request, starts a microsecond timer, records multi-hop spans across REST, gRPC, and Kafka, and streams traces asynchronously to `request-tracker-service` via Kafka (`http-request-traces`).
- **Distributed Correlation**: Injects `x-request-id` and `x-trace-id` into downstream gRPC context metadata and Kafka event headers.
- **Client Authentication**: Issues and validates HS256 JWT tokens for browser sessions and developer APIs.
- **Service-to-Service Security**: Injects `X-Service-Name: api-gateway` and cryptographic `Authorization: Bearer <AUTH_TOKEN>` on all outgoing gRPC calls.
- **Built-in Mock Webhook Receiver**: `POST /api/v1/webhooks/test-receiver` captures HMAC signatures and incoming payloads for local testing with zero external dependencies.
- **Flexible Dispatching**: `POST /api/v1/webhooks/send` and `GET /api/v1/webhooks/send` accept parameters both via JSON body and URL query parameters (`app_id`, `event_name`, `target_url_override`, `payload`).
- **Microservices Health Monitoring**: `/health` endpoint checks gateway status and pings upstream gRPC microservices with latency reporting.

---

## 🚀 Running the API Gateway

### 1. Configuration
Copy the sample environment file:
```bash
cp .env.example .env
```

Ensure `ACCOUNTS_GRPC_PORT=50051`, `SUBSCRIPTIONS_GRPC_PORT=50052`, `RUNNER_GRPC_PORT=50053`, and `AUTH_TOKEN` match your microservices settings.

### 2. Run the Gateway
```bash
go run cmd/server/main.go
```
The server will start on `http://localhost:8080`.

---

## 📡 Key REST Endpoints

### 1. Webhooks & Applications
| Method | Path | Description | Access |
|---|---|---|---|
| `POST` | `/api/v1/webhooks/test-receiver` | Built-in local mock webhook receiver | Public |
| `POST` / `GET` | `/api/v1/webhooks/send` | Dispatch webhook (supports JSON & query params) | JWT |
| `GET` | `/api/v1/webhooks/calls` | List execution traces (supports `app_id` filter) | JWT |
| `GET` | `/api/v1/webhooks/calls/:id` | Get single webhook call details | JWT |
| `POST` | `/api/v1/webhooks/calls/:id/retry` | Retry failed webhook call | JWT |
| `GET` | `/api/v1/apps` | List registered applications | JWT |
| `POST` | `/api/v1/apps` | Create new application | JWT |
| `PUT` | `/api/v1/apps/:id` | Update application | JWT |
| `DELETE` | `/api/v1/apps/:id` | Delete application | JWT |
| `POST` | `/api/v1/apps/:id/rotate-secrets` | Rotate app client secret / HMAC key | JWT |

### 2. Subscriptions & Billing
| Method | Path | Description | Access |
|---|---|---|---|
| `GET` | `/api/v1/plans` | List available subscription tiers | Public |
| `GET` | `/api/v1/subscriptions/my` | Get current user's active subscription | JWT |
| `POST` | `/api/v1/subscriptions/subscribe` | Subscribe to plan | JWT |
| `POST` | `/api/v1/subscriptions/cancel` | Cancel subscription | JWT |
| `GET` | `/api/v1/invoices` | List invoices | JWT |
| `POST` | `/api/v1/invoices/:id/manual-payment` | Submit bank wire transaction reference | JWT |
| `GET` | `/api/v1/admin/manual-payments` | Admin offline payment review queue | Admin JWT |
| `POST` | `/api/v1/admin/manual-payments/:id/review` | Approve/Reject bank wire payment | Admin JWT |

### 3. Accounts & RBAC
| Method | Path | Description | Access |
|---|---|---|---|
| `GET` | `/health` | Live gateway & gRPC health ping | Public |
| `POST` | `/api/v1/auth/login` | User/Admin authentication | Public |
| `POST` | `/api/v1/auth/register` | Self-service registration | Public |
| `GET` | `/api/v1/auth/me` | Authenticated profile info | JWT |
| `GET` | `/api/v1/users` | List users (paginated + search) | JWT |
| `POST` | `/api/v1/users` | Create user | JWT |
| `PUT` | `/api/v1/users/:id` | Update user | JWT |
| `DELETE` | `/api/v1/users/:id` | Delete user | JWT |
| `GET` | `/api/v1/admins` | List administrators | JWT |
| `GET` | `/api/v1/roles` | List roles | JWT |
| `GET` | `/api/v1/permissions` | List system permissions | JWT |

### 4. Request Traces & System APM Telemetry
| Method | Path | Description | Access |
|---|---|---|---|
| `GET` | `/api/v1/request-traces` | List request traces with multi-filter (actor, route, status, latency) | Admin JWT |
| `GET` | `/api/v1/request-traces/:id` | Get single request trace snapshot with payload and trip waterfall | Admin JWT |
| `GET` | `/api/v1/request-traces/stats` | Aggregated latency metrics (total, avg ms, P95/P99, error rate %) | Admin JWT |

### 5. Audit Logging & Compliance
| Method | Path | Description | Access |
|---|---|---|---|
| `GET` | `/api/v1/audit-logs` | List immutable audit log entries (actor, action, resource, diffs) | Admin JWT |
| `GET` | `/api/v1/audit-logs/:id` | Get single audit log event details | Admin JWT |
