# Page Specification: Provider Network (`/providers`)

## Purpose
Manage upstream AI backends (OpenAI, Anthropic Claude, Google Gemini, Groq, Cerebras, OpenRouter, Cloudflare AI, Ollama, Custom backends) and their configured accounts.

## Route
- List: `/providers`
- Create Provider: `/providers/new`
- Detail & Accounts: `/providers/[id]`

## Access Requirement
Admin role.

## User Goals
- View all upstream providers, their kinds, base URLs, and active account counts.
- Add a new AI provider endpoint without using a modal.
- Inspect and manage accounts, API keys, and priority levels for each provider.

## Page Layout & Sections
- **Header**: Title *"AI Provider Network"*, Description *"Manage upstream model backends and connection endpoints."*, Primary Action button `[+ Add Provider]` linking to `/providers/new`.
- **Filter Toolbar**: Search by provider name/key; filter by kind dropdown.
- **Data Table**:
  - Columns: Provider Name, Kind Badge, Upstream URL, Active Accounts, Status Toggle, Actions (`Inspect`, `Delete`).
- **Provider Detail View (`/providers/[id]`)**:
  - Top Card: Provider Metadata, Base URL, Health Status.
  - Accounts Section: Table of accounts with Name, Priority, State (`active` / `cooling_down`), and inline `[+ Add Account]` expandable form.

## Actions & No-Modal Pattern
- **Create Provider**: Navigates to dedicated page `/providers/new` with full form (ID, Key, Name, Kind, Base URL, Enabled).
- **Add Account**: Expands an inline collapsible card above the accounts table on `/providers/[id]`.
- **Delete Provider**: Uses `InlineConfirm` inside the table row.

## Data Endpoints
- `GET /api/providers`, `POST /api/providers`, `DELETE /api/providers/{id}`.
- `GET /api/accounts`, `POST /api/accounts`, `DELETE /api/accounts/{id}`.

## Required SVG Icons
- `Cloud`, `Server`, `Plus`, `Trash2`, `Edit`, `ExternalLink`, `ShieldCheck`, `AlertCircle`.
