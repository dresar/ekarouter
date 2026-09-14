# 08. Master Page Specifications & UX Layouts

This document establishes the universal page layout structure and functional requirements across the EkaRouter console. For granular endpoint mappings and field-level validation schemas, refer to the corresponding documents in [`25_PAGE_DETAILS/`](./25_PAGE_DETAILS/).

---

## 1. Master Page Shell Layout

Every authenticated page renders inside the master `AppShell`:

```text
┌───────────────────────────────────────────────────────────────────────────┐
│ Topbar: System Health Pill | Active Route | Theme Toggle | User Avatar    │
├──────────────┬────────────────────────────────────────────────────────────┤
│ Sidebar      │ Main Content Container                                     │
│ (240px wide) │ ┌────────────────────────────────────────────────────────┐ │
│              │ │ PageHeader: Title, Breadcrumbs, Primary Action Buttons │ │
│ - Gateway    │ ├────────────────────────────────────────────────────────┤ │
│ - Platform   │ │ Filter / Search Bar (Dense, 36px high)                 │ │
│ - Access     │ ├────────────────────────────────────────────────────────┤ │
│ - Settings   │ │ Main Content: Tables, Inline Panels, Compact Cards     │ │
│              │ └────────────────────────────────────────────────────────┘ │
└──────────────┴────────────────────────────────────────────────────────────┘
```

---

## 2. Key Pages Summary

### 1. Command Center (`/overview`)
- **Purpose**: Real-time operational dashboard for gateway traffic and health.
- **Key Sections**:
  - Top Metrics Ribbon: Active Providers, Total Requests, Cached Token Ratio, Active Cooldowns.
  - Active Failover Status: Providers currently in backoff with countdown timers.
  - Recent Gateway Activity: Live request timeline with latency and status badges.
  - Quick Ingress Test: Interactive curl / fetch snippet generator.

### 2. Provider Management (`/providers`)
- **Purpose**: Configure upstream AI backends and associated account credentials.
- **Key Sections**:
  - Providers Data Table: Name, kind badge, upstream base URL, active accounts count, status toggle.
  - Detail View (`/providers/[id]`): Multi-account credentials list with priority and cooldown state.
  - Non-Modal Pattern: New account form renders as an inline collapsible card or full page (`/providers/new`).

### 3. Routing & Combos (`/routing`)
- **Purpose**: Manage model routing combos, fallbacks, and rotation algorithms.
- **Key Sections**:
  - Route Card Grid: Model slug (e.g. `gpt-4o`), strategy indicator (`priority`, `round_robin`), and ordered target accounts.
  - Target Reordering: Priority sequence display with drag-or-number priority order.
  - Creation / Edit: Dedicated page (`/routing/new`) with target account picker.

### 4. Credential Vault (`/vault`)
- **Purpose**: Secure developer secret management with masking and rotation.
- **Key Sections**:
  - Vault Table: Name, provider category, environment tag (`prod`/`staging`), masked secret (`sk-****abcd`), health badge.
  - Connectivity Tester: Inline trigger to test credential against provider API.
  - Rotate Secret: Dedicated side-drawer or detail view to trigger secret update.

### 5. Outbound Proxies (`/proxies`)
- **Purpose**: Outbound routing for geographic bypass or rate-limit segregation.
- **Key Sections**:
  - Proxy Profile List: Scheme (`http`, `socks5`, `relay`), host, port, authentication status.
  - Test Action: Live HTTP latency check with response time indicator.

### 6. Ingress API Keys (`/api-keys`)
- **Purpose**: Authorize client applications to call `/v1/*` endpoints.
- **Key Sections**:
  - Key List: Prefix, name, scopes, created date, last used timestamp, revoke button.
  - Key Generation Flow: Displays full generated key once with copy-to-clipboard confirmation.

### 7. Usage & Telemetry (`/usage`)
- **Purpose**: Review historical token consumption and provider latency.
- **Key Sections**:
  - Token Breakdown: Prompt tokens vs. completion tokens.
  - Provider Distribution: Share of requests across OpenAI, Anthropic, Gemini, etc.
  - Filtering: Quick range selector (Today, 7 Days, 30 Days).

### 8. System Settings & Backups (`/settings` & `/backup`)
- **Purpose**: Control server configuration and execute database maintenance.
- **Key Sections**:
  - Environment Variables Inspection: Read-only view of runtime environment.
  - SQLite Backup Management: List existing `.db` backup archives and trigger online vacuum backup.
