# Page Specification: Quotas & Rate Limits (`/quota`)

## Purpose
Monitor rate-limit buckets, account cooldowns, and active backoff timers across all AI backends.

## Route
`/quota` (or accessed via `/providers` account detail tab).

## Access Requirement
Admin role.

## User Goals
- Identify accounts currently blocked by rate limit cooldowns.
- View countdown timer until an account is eligible to resume traffic.
- Inspect historical rate limit triggers and error counts.

## Page Layout
- **Header**: Title *"Account Quotas & Cooldown Circuits"*.
- **Active Cooldowns Banner**:
  - Highlights accounts currently in backoff with remaining countdown time.
  - Quick action: "Reset Cooldown" (resets state to active).
- **Accounts Quota Table**:
  - Columns: Account Name, Provider, State Badge (`Active` / `Cooling Down`), Failure Count, Cooldown Expiry, Actions.

## Data Endpoints
- `GET /api/accounts`, `POST /api/accounts`.

## Required SVG Icons
- `Gauge`, `Clock`, `AlertTriangle`, `RotateCcw`, `ShieldAlert`.
