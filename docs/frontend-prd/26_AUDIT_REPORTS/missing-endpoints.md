# Audit Report: Missing Endpoints & Deprecations

Inventory of endpoints that are conceptually discussed in developer platform documentation or third-party AI gateways but do not exist in the EkaRouter Go backend.

---

## 1. Missing Endpoints Matrix

| Intended Route | Method | Classification | Backend Analysis | Frontend Policy |
| :--- | :--- | :--- | :--- | :--- |
| `/v1/messages` | `POST` | `NOT_IMPLEMENTED` | Native Anthropic ingress format is not yet registered on `apiV1` chi router. | Do not implement in frontend. All Anthropic models route through `/v1/chat/completions`. |
| `/v1/embeddings` | `POST` | `NOT_IMPLEMENTED` | Vector embeddings handler is not registered. | Omit embedding UI controls from console. |
| `/api/logs/stream` | `WS` | `NOT_IMPLEMENTED` | No WebSocket server is initialized in `internal/httpapi`. | Poll `/api/v1/audit-logs` at 10-second intervals. |
| `/api/auth/reset-password` | `POST` | `NOT_IMPLEMENTED` | No email provider or reset token table exists. | Provide static guidance to reset admin credentials via CLI. |
| `/api/billing/invoices` | `GET` | `NOT_IMPLEMENTED` | No financial accounting ledger exists in SQLite schema. | Only display raw token usage telemetry with client-side estimates. |

---

## 2. Deprecations & Unused Routes

- `/api/credentials` (legacy endpoint): Superseded by `/api/v1/credentials` (Vault API). Frontend should prioritize the `/api/v1/credentials` endpoint for all developer secret operations.
