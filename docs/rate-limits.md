# Rate Limiting & Cooldown Management

## Sliding Window Rate Limiter
EkaRouter maintains an in-memory sliding-window timestamp buffer per client/project key.
Requests exceeding the rate limit receive HTTP 429 Too Many Requests with a `Retry-After` header.

## Cooldown Management
When an upstream provider returns HTTP 429 or rate-limit signals:
1. The credential is immediately placed into cooldown with `cooldown_until = now + RetryAfter`.
2. A record is inserted into `credential_cooldowns` with the failure reason.
3. The rotater falls back to the next eligible credential seamlessly.
4. When `cooldown_until` expires, the background scheduler or rotator automatically marks the key available.
