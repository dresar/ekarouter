# 20. Backend Gaps, Missing Endpoints & Enhancements

This document catalogues items where frontend requirements exceed current backend implementations. The frontend AI agent must treat these as non-negotiable boundaries and not hallucinate client-side workarounds.

---

## 1. Identified Backend Gaps

| Feature Area | Missing Backend Capability | Current Workaround in Frontend | Future Backend Recommendation |
| :--- | :--- | :--- | :--- |
| **Realtime Logs** | No WebSocket endpoint for live log streaming. | Poll `GET /api/v1/audit-logs` every 10 seconds. | Implement WebSocket or SSE log stream at `/api/v1/logs/stream`. |
| **Cost Invoicing** | Backend tracks token counts, not currency costs. | Calculate estimated costs on client with explicit "Estimate" label. | Add model pricing table in SQLite database. |
| **Direct Messages**| No `/v1/messages` native Anthropic endpoint. | Route Anthropic requests through `/v1/chat/completions` translation. | Add native `/v1/messages` handler in `internal/httpapi`. |
| **Vector Search** | No `/v1/embeddings` endpoint. | Omit vector embedding controls from frontend. | Implement embedding provider adapters. |
| **Password Reset**| No self-service email password reset. | Display notice: *"Contact server admin to reset password via CLI."* | Add email transport or CLI reset subcommand. |

---

## 2. Invariant Rule for Frontend Engineers

Under no circumstances should the frontend:
- Attempt to connect to `ws://...` or `wss://...` endpoints until a WebSocket server is implemented in Go.
- Create mock routes that masquerade as real backend endpoints.
- Invent fake billing or credit card processing screens.
