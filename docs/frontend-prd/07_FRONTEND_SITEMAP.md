# 07. Frontend Information Architecture & Sitemap

---

## 1. Top-Level Navigation Structure

The EkaRouter console navigation is divided into 5 functional areas in the left sidebar:

```text
├── PUBLIC
│   ├── /login                     (Sign in screen)
│   ├── /404                       (Not Found)
│   └── /500                       (Server Error / Offline)
│
├── AI GATEWAY
│   ├── /overview                  (Command Center & Gateway Telemetry)
│   ├── /providers                 (AI Providers & Upstream Endpoints)
│   │   ├── /providers/new         (Create Provider)
│   │   └── /providers/[id]        (Provider Detail & Accounts)
│   ├── /routing                   (Routing Combos & Fallback Targets)
│   │   ├── /routing/new           (Create Route)
│   │   └── /routing/[id]          (Route Detail & Strategy)
│   ├── /models                    (Model Aliases & Capabilities)
│   └── /token-saver               (Prompt Compaction Playground)
│
├── DEVELOPER PLATFORM
│   ├── /vault                     (Encrypted Credentials & Health)
│   │   ├── /vault/new             (Add Vault Secret)
│   │   └── /vault/[id]            (Credential Detail & Rotation History)
│   ├── /tools                     (Generic Tools & Request Templates)
│   │   ├── /tools/new             (Define Tool)
│   │   └── /tools/[id]            (Tool Schema & Execution Test)
│   ├── /proxies                   (Outbound Proxy Profiles)
│   │   └── /proxies/new           (Add Proxy Profile)
│   └── /free-tiers                (Discovered Community Free Tiers)
│
├── ACCESS & OBSERVABILITY
│   ├── /api-keys                  (Ingress Client API Keys)
│   ├── /usage                     (Token Consumption & Activity)
│   └── /audit-logs                (Immutable System Audit Trail)
│
└── SETTINGS
    ├── /settings                  (System Configuration & Environment)
    ├── /backup                    (Database Snapshot & Backups)
    └── /api-docs                  (Interactive OpenAPI Explorer)
```

---

## 2. Page & Permission Map

| Route | Nav Group | Required Role | Primary Backend Endpoints |
| :--- | :--- | :--- | :--- |
| `/login` | Public | None | `POST /api/auth/login` |
| `/overview` | AI Gateway | Any | `GET /health`, `GET /api/usage`, `GET /api/v1/health` |
| `/providers` | AI Gateway | Admin | `GET /api/providers`, `GET /api/accounts` |
| `/providers/[id]` | AI Gateway | Admin | `GET /api/providers`, `POST /api/accounts` |
| `/routing` | AI Gateway | Admin | `GET /api/routes`, `POST /api/routes` |
| `/routing/[id]` | AI Gateway | Admin | `GET /api/routes`, `DELETE /api/routes/{id}` |
| `/models` | AI Gateway | Admin | `GET /v1/models`, `GET /api/models` |
| `/token-saver` | AI Gateway | Any | `POST /api/tokensaver/preview` |
| `/vault` | Developer | Admin/Dev | `GET /api/v1/credentials`, `GET /api/v1/health/credentials` |
| `/vault/[id]` | Developer | Admin/Dev | `GET /api/v1/credentials/{id}`, `POST /rotate` |
| `/tools` | Developer | Admin/Dev | `GET /api/v1/tools`, `GET /api/v1/request-templates` |
| `/tools/[id]` | Developer | Admin/Dev | `GET /api/v1/tools/{id}`, `POST /execute` |
| `/proxies` | Developer | Admin | `GET /api/proxies`, `POST /api/proxies/{id}/test` |
| `/free-tiers` | Developer | Any | `GET /api/free-tiers`, `GET /api/free-tiers/verified` |
| `/api-keys` | Access | Admin | `GET /api/keys`, `POST /api/keys`, `DELETE /api/keys/{id}` |
| `/usage` | Access | Any | `GET /api/usage`, `GET /api/v1/usage/summary` |
| `/audit-logs` | Access | Admin | `GET /api/v1/audit-logs` |
| `/settings` | Settings | Admin | `GET /api/settings`, `PUT /api/settings` |
| `/backup` | Settings | Admin | `GET /api/backup`, `POST /api/backup` |
| `/api-docs` | Settings | Any | Static schema / `GET /ready` |
