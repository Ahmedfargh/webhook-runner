# Accounts Service

High-performance gRPC microservice for account, identity, role, and permission management built with Golang, GORM, Protobuf, and the HMVC (Hierarchical Model-View-Controller) architecture pattern.

---

## 🏛 HMVC Architecture

Each domain feature is organized as a self-contained module in `internal/modules/<feature>`:

- **Model**: Database entities ([internal/models](file:///home/ahmed/golang/webhook-project/accounts/internal/models)) and GORM repository interface + implementation (`repository/`).
- **View**: Protobuf message mappers and response serializers (`presenter/`).
- **Controller**: gRPC handlers implementing the generated service contracts (`controller/`).
- **Service**: Business logic validation, phone normalization, and password hashing (`service/`).
- **Module Orchestrator**: Dependency injection and gRPC registration (`<feature>_module.go`).

---

## 🔐 Token-Based Authentication

All gRPC RPC methods (except reflection) require a Bearer token passed via gRPC metadata headers:
```
authorization: Bearer <AUTH_TOKEN>
x-service-name: api-gateway
x-request-id: <correlation-id>
x-trace-id: <correlation-id>
```

All incoming calls have their `x-request-id` and `x-trace-id` extracted for logging and audit correlation. High-impact operations (login, user registration, role assignments, permissions changes) are emitted to Kafka topic `audit-events`.

### Generate Auth Token Command

To generate a cryptographically secure token and automatically save it to your `.env` file:

```bash
go run cmd/token/main.go
```

#### Flags:
- `--save=true|false`: Automatically update or append `AUTH_TOKEN` in `.env` (default: `true`).
- `--length=32`: Number of entropy bytes (default: 32 bytes = 64 hex characters).
- `--env=.env`: Path to the `.env` file.

---

## 🚀 Running the Service

### 1. Migrations & Seeders
```bash
go run cmd/migrations/migrate.go
go run cmd/seeders/Seeder.go
```

### 2. Start the gRPC Server
```bash
go run cmd/server/main.go
```
The server listens on `50051` (or `GRPC_PORT` from `.env`) and supports gRPC reflection.

---

## 🧪 Testing

Run the test suite:
```bash
# Run all module tests
go test -v ./internal/modules/...

# Run middleware tests
go test -v ./internal/middleware/...

# Run helper tests
go test -v ./internal/helpers/...
```

---

## 📮 Postman Collection

Import the pre-configured Postman collection located at:
`accounts/api/postman/accounts_grpc_collection.json`

It includes:
- Collection variables: `{{grpc_host}}`, `{{grpc_port}}`, and `{{auth_token}}`.
- Metadata authorization header pre-set on every request.
- Sample requests and payloads for **Users**, **Admins**, **Roles**, and **Permissions**.
