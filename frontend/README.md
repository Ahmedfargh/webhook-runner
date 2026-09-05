# Webhook Platform Frontend (Vue 3 + Vite)

Enterprise single-page application built with Vue 3, Vite, Pinia, Vue Router, and Lucide Icons, featuring **Zoho Projects** aesthetics, comprehensive 6-language internationalization (Arabic, English, French, German, Russian), and full RBAC views.

---

## ✨ Key Features & Views

### 1. Applications & Webhook Runner (`/apps`)
- **App Credentials Management**: View and rotate Client ID, Client Secret, and Webhook HMAC Secrets.
- **Dedicated Webhook Logs Inspector**: Click **`Webhook Logs`** on any app to open real-time delivery logs, success rate %, latency benchmarks, and payload inspectors.
- **Instant Test Webhook Dispatcher**: 1-click **`⚡ Use Local Receiver`** shortcut pointing to the built-in mock receiver (`http://localhost:8080/api/v1/webhooks/test-receiver`).

### 2. Multi-Language Internationalization (i18n)
- **6 Locales Supported**: Arabic (العربية - RTL), English, French (Français), German (Deutsch), Russian (Русский).
- Instant locale switching with persistence across reloads.

### 3. Subscriptions & Billing
- **Plans & Pricing (`/plans`)**: Browse tiered subscription tiers (Free, Starter, Pro, Enterprise) with annual discount toggles.
- **My Subscription (`/subscription`)**: Manage plan status, renewal dates, quota consumption, and bank wire payment details.
- **Invoices Ledger (`/invoices`)**: View invoices, print tax receipts, and submit bank transfer wire reference numbers.
- **Admin Billing Console (`/admin/billing`)**: Review offline bank transfer payment queue, approve/reject wire proofs, issue manual custom invoices, and perform admin subscription overrides.

### 4. Identity & RBAC Management (Admin Only)
- User Management (`/users`), Admin Accounts (`/admins`), Roles (`/roles`), and Granular Permission Matrix (`/permissions`).

### 5. System Observability & Topology (`/topology`)
- Live microservices visualizer showing Browser $\rightarrow$ API Gateway $\rightarrow$ gRPC Services $\rightarrow$ MySQL.

---

## 🚀 Running the Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`.

---

## 🧪 Production Build

```bash
npm run build
```
Generates production-optimized bundle in `dist/`.
