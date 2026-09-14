# Endpoint Details: Quotas & Rate Limits

---

## 1. Rate Limiter Architecture

EkaRouter implements sliding-window rate limiting in `internal/limits/limits.go`:
- Reference Types: `api_key`, `account`, `project`, `ip`.
- Sliding Window: 60-second default.
- Storage: In-memory sliding window backed by SQLite `quota_records`.

---

## 2. Quota Telemetry via `/api/accounts`

- **Purpose**: Accounts list returns live quota state and active cooldown expiry.
- **Account State Schema**:
  ```json
  {
    "id": "acc_openai_1",
    "provider_id": "prov_openai",
    "name": "Production Tier 1",
    "state": "cooling_down",
    "cooldown_until": "2026-09-15T05:02:15Z",
    "failure_count": 3
  }
  ```
- **Frontend Presentation**:
  - `state == 'active'`: Green badge.
  - `state == 'cooling_down'`: Amber badge with live countdown timer to `cooldown_until`.

---

## 3. Cooldown Reset & Manual Recovery

- Accounts recover automatically when `time.Now().After(cooldown_until)`.
- Re-enabling or editing an account resets its cooldown state to `active`.
