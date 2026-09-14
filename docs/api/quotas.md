# EkaRouter Quotas & Rate Limits API

## 1. Overview

EkaRouter tracks rate limits and usage quotas per client API key, per project, and per upstream provider credential.

---

## 2. Limits & Quotas Engine

1. **Sliding Window Rate Limiting**:
   - Monitored by `limits.Engine` in `internal/limits/`.
   - Tracks requests per minute (`RPM`) and tokens per minute (`TPM`).
   - If a client exceeds their configured rate limit, EkaRouter returns HTTP 429 Too Many Requests with a `Retry-After` header.

2. **Upstream Quota Tracking**:
   - Monitored via `quota_snapshots` in the database.
   - When an upstream provider returns HTTP 402 or an error indicating insufficient quota, the account enters cooldown and fallback triggers.

---

## 3. Querying Quotas and Limits

- Platform usage endpoints (`GET /api/v1/usage/summary`) return active quota snapshots and limits.
- The UI can display remaining token and request allowances per billing interval.
