# Subscriptions & Billing Microservice

Enterprise-grade subscription lifecycle, multi-tier pricing, tax-compliant invoicing, and offline bank wire payment processing service built in Go, GORM, Protobuf, and HMVC architecture.

---

## 🏛 HMVC Architecture

Organized into domain modules inside `internal/modules/`:

```
subscriptions/
├── api/
│   └── proto/v1/                # Generated gRPC Protobuf stubs
├── cmd/
│   ├── server/                  # Service entrypoint (Port 50052)
│   ├── migrations/              # GORM auto-migrations
│   └── seeders/                 # Realistic subscription tiers & seed data
└── internal/
    ├── config/                  # Database connection pool & environment
    ├── middleware/              # Service auth & token interceptors
    ├── models/                  # GORM entities (Plan, Subscription, Invoice, ManualPayment)
    └── modules/
        ├── plan/                # Pricing tiers, feature quotas, limits
        ├── subscription/        # Lifecycle: subscribe, auto-provision free, cancel, admin override
        ├── invoice/             # Automated invoice generation, numbering, tax computation
        └── manual_payment/      # Bank wire proof upload, review queue, auto-activation
```

---

## 📊 Database Entities (`webhook_subscriptions`)

### 1. `plans`
- Multi-tier plans (`Free`, `Starter`, `Pro`, `Enterprise`).
- `monthly_price`, `annual_price`, `currency`.
- Quotas: `max_apps`, `monthly_events_limit`, `rate_limit_per_second`.
- `features_json`: Rich feature list for frontend pricing display.

### 2. `subscriptions`
- `user_id`: Reference to the customer account.
- `plan_id`: Subscribed tier.
- `status`: `ACTIVE`, `CANCELLED`, `PAST_DUE`, `TRIALING`.
- Billing period tracking: `current_period_start`, `current_period_end`.

### 3. `invoices`
- Tax-compliant numbering (`INV-YYYYMM-XXXX`).
- `subtotal`, `tax_rate`, `tax_amount`, `total_amount`.
- Payment status: `PAID`, `PENDING`, `OVERDUE`, `VOID`.

### 4. `manual_payments`
- Bank transfer / wire reference tracker.
- `transaction_reference`: Wire reference code provided by customer.
- `proof_image_path`: Uploaded payment receipt.
- `status`: `PENDING`, `APPROVED`, `REJECTED`.
- Approving an offline payment automatically marks the invoice as `PAID` and activates the subscription.

---

## 📡 gRPC Services Definition

Defined in `api/proto/v1/`:
- `PlanService`: `ListPlans`, `GetPlan`, `CreatePlan`, `UpdatePlan`, `DeletePlan`
- `SubscriptionService`: `GetMySubscription`, `Subscribe`, `CancelSubscription`, `AdminOverrideSubscription`
- `InvoiceService`: `ListInvoices`, `GetInvoice`, `CreateManualInvoice`
- `ManualPaymentService`: `SubmitManualPayment`, `ListPendingPayments`, `ApproveManualPayment`, `RejectManualPayment`

---

## 🔐 Service Security & Audit Trail

- Enforces gRPC metadata authentication (`Authorization: Bearer <AUTH_TOKEN>`) and caller whitelisting (`ALLOWED_SERVICES=api-gateway`).
- Asynchronously emits events to Kafka topic `audit-events` on plan modifications, subscription cancellations, invoice creations, and manual wire approvals.

---

## 🚀 Running the Service

### 1. Configuration
Environment variables in `.env`:
```env
DB_HOST=127.0.0.1
DB_PORT=3307
DB_USER=root
DB_PASSWORD=secret
DB_NAME=webhook_subscriptions
GRPC_PORT=50052
AUTH_TOKEN=webhook-accounts-secret-token
ALLOWED_SERVICES=api-gateway,webhook-runner
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_AUDIT=audit-events
```

### 2. Run Locally
```bash
cd subscriptions
go run cmd/migrations/main.go
go run cmd/seeders/main.go
go run cmd/server/main.go
```
Listens on gRPC port `50052`.
