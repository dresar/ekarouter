# Limits Package

## Purpose
The `limits` package provides rate limiting, quota tracking, and cooldown management across API keys, providers, projects, and users.

## Responsibilities
- Concurrency-safe sliding window rate limiting in memory.
- Persistent SQLite quota records with standard quota states (`unknown`, `available`, `limited`, `exhausted`, `suspended`, `invalid`, `expired`, `disabled`).
- Credential cooldown recording upon HTTP 429 or provider quota exhaustion with persisted `ends_at` timestamps.

## Public Interfaces
- `NewEngine(db *sql.DB) *Engine`
- `(*Engine) AllowRate(key string, limit int, window time.Duration) bool`
- `(*Engine) SetQuota(ctx context.Context, refType, refID, metric, period string, maxVal int64, resetAt *time.Time) error`
- `(*Engine) CheckQuota(ctx context.Context, refType, refID, metric string) (QuotaState, int64, int64, error)`
- `(*Engine) IncrementQuota(ctx context.Context, refType, refID, metric string, delta int64) error`
- `(*Engine) SetCooldown(ctx context.Context, credID string, duration time.Duration, reason string) error`
- `(*Engine) IsInCooldown(ctx context.Context, credID string) (bool, time.Duration, error)`

## Security Considerations
- Quota and cooldown calculations are deterministic and protected against concurrent race conditions.
- Error codes and cooldown states do not leak credential secrets.
