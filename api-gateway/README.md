# Webhook API Gateway

High-performance REST API Gateway written in Go (Gin) for orchestrating microservices, handling client authentication (JWT), and securely forwarding requests to internal services via gRPC.

---

## 🏛 Architecture & Security Features

- **Public REST API**: Exposes clean JSON REST endpoints for Users, Admins, Roles, Permissions, Auth, and Health.
- **Client Authentication**: Issues and validates HS256 JWT tokens for browser & external client sessions.
- **Service-to-Service Security**: Injects `X-Service-Name: api-gateway` and cryptographic `Authorization: Bearer <AUTH_TOKEN>` on all outgoing gRPC calls to the Accounts Service.
- **CORS Support**: Preconfigured CORS middleware for cross-origin integration with Vue.js frontend.
- **Microservices Health Monitoring**: `/health` endpoint checks API Gateway status and pings downstream gRPC microservices with latency reporting.

---

## 🚀 Running the API Gateway

### 1. Configuration
Copy the sample environment file:
```bash
cp .env.example .env
```

Ensure `ACCOUNTS_GRPC_HOST`, `ACCOUNTS_GRPC_PORT`, and `SERVICE_TOKEN` match your Accounts service settings.

### 2. Run the Gateway
```bash
go run cmd/server/main.go
```
The server will start on `http://localhost:8080`.

---

## 📡 Key Endpoints

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
| `POST` | `/api/v1/admins/:id/roles` | Assign roles to admin | JWT |
| `GET` | `/api/v1/roles` | List roles | JWT |
| `POST` | `/api/v1/roles/:id/permissions` | Assign permissions to role | JWT |
| `GET` | `/api/v1/permissions` | List system permissions | JWT |
