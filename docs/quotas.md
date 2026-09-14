# Quotas & Usage Limits

EkaRouter tracks request and token quotas across credentials, projects, and providers.

## Standard Quota States

| State | Definition |
|---|---|
| `unknown` | Quota is unmetered or not reported by the upstream API. |
| `available` | Usage is below 90% of configured threshold. |
| `limited` | Usage has reached or exceeded 90% of threshold; warning emitted. |
| `exhausted` | Usage has reached 100%; further requests trigger cooldown or fallback. |
| `suspended` | Upstream account is flagged or temporarily blocked. |
| `invalid` | Key authentication was rejected upstream (HTTP 401/403). |
| `expired` | Token expiration date has passed. |
| `disabled` | Manually deactivated by administrator. |

## Reset Cycles
Quotas support daily, weekly, and monthly periods. Upon reaching the `reset_at` timestamp, the scheduler or engine resets the `used_value` counter to zero.
