# API Keys & Client Tokens

EkaRouter distinguishes between two categories of internal API tokens:

## 1. Gateway API Keys (`eka_live_*`)
- **Prefix:** `eka_live_` (e.g. `eka_live_a1b2c3d4e5f6...`).
- **Target:** Universal AI Gateway endpoints (`/v1/chat/completions`, `/v1/models`, `/v1/responses`).
- **Management:** Created via `POST /api/keys`.
- **Storage:** SHA-256 hashed. Plaintext is only displayed once upon generation.

## 2. Personal Access Tokens (`eka_pat_*`)
- **Prefix:** `eka_pat_` (e.g. `eka_pat_f9e8d7c6b5a4...`).
- **Target:** Unified Developer Platform endpoints (`/api/v1/*`).
- **Scopes:** Scoped to specific projects, environments, or permissions.
- **Expiration:** Configurable TTL (e.g., 30, 90, 365 days).
- **Storage:** SHA-256 hashed in `client_tokens`.
