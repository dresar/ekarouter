# Page Specification: Ingress API Keys (`/api-keys`)

## Purpose
Manage client access tokens used by external development tools (Cursor, Aider, OpenCode, Claude CLI) to authenticate against the EkaRouter `/v1/*` gateway.

## Route
`/api-keys`

## Access Requirement
Admin role.

## User Goals
- View active API keys, prefixes, and last used timestamps.
- Generate a new client API key with custom scopes.
- Copy the newly generated secret key immediately.
- Revoke compromised or obsolete API keys.

## Page Layout
- **Header**: Title *"Ingress API Keys"*, Description *"Client tokens for accessing EkaRouter OpenAI-compatible gateway."*, Primary button `[+ Generate Key]`.
- **Active Key Notification Banner**: When a key is created, renders a prominent high-contrast banner containing the unmasked key, copy button, and a mandatory *"I have saved this key"* acknowledgment button.
- **Key List Table**:
  - Columns: Key Name, Token Prefix (`eka_live_...`), Scopes, Created Date, Last Used At, Status Badge, Revoke Action.

## Security Considerations
- The full token is never visible after the initial creation banner is dismissed.
- Copy action triggers visual confirmation (`Check` icon for 2 seconds).

## Data Endpoints
- `GET /api/keys`, `POST /api/keys`, `DELETE /api/keys/{id}`.

## Required SVG Icons
- `Key`, `Copy`, `Check`, `AlertTriangle`, `Trash2`, `Lock`.
