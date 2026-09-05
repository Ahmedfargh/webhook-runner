# Webhook Accounts Management Portal (Vue 3)

Enterprise single-page application built with Vue 3, Vite, Pinia, Vue Router, and Lucide Icons, designed with **Zoho Projects** aesthetics and clean SOLID architecture.

---

## ✨ Features

- **Zoho Projects Enterprise Design**:
  - Deep slate collapsible navigation sidebar.
  - Interactive breadcrumbs and global `/` search bar shortcut.
  - Slide-over drawers for creating and updating resources without leaving context.
  - Dynamic status pills, filter chips, and pagination controls.
  - Floating toast notification system.
- **Unified Microservices Observability**:
  - Live **Gateway Topology** monitor visualizing communication flows between Browser -> API Gateway (REST + JWT) -> Accounts Service (gRPC + Caller Service Authorization) -> Database.
  - Real-time gRPC connectivity and latency diagnostics.
- **SOLID & Clean Architecture**:
  - Modular service layer (`userService`, `adminService`, `roleService`, `permissionService`, `healthService`).
  - Pinia stores for authentication and global notifications.
  - Highly reusable atomic base components (`Drawer`, `Modal`, `StatCard`, `Badge`, `Pagination`, `EmptyState`).

---

## 🚀 Running the Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`.
