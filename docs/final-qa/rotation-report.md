# EkaRouter Rotation Engine Test Report

**Audit Date:** 2026-09-15  
**Component:** `internal/rotator` and `internal/credpool`  
**Test Suite:** `TestRotatorStrategies`, `TestConcurrentRotatorSafety`

---

## 1. Supported Rotation Strategies

| Strategy | Sorting / Selection Rule | Verified Scenario | Result |
|---|---|---|---|
| **Priority** (`priority`) | Lowest priority integer rank (`priority ASC`), tie-broken by `request_count ASC` | Primary key with Priority 5 selected over Priority 10 | **PASS** |
| **Least Used** (`least_used`) | Candidate with lowest total `request_count ASC` | Key with 5 requests selected over key with 10 requests | **PASS** |
| **Round Robin** (`round_robin`) | Atomic round-robin cursor indexed modulo candidate count | Even sequential distribution across 3 keys | **PASS** |
| **Random** (`random`) | Uniform pseudorandom integer selection | Random distribution across pool | **PASS** |
| **Health Based** (`health_based`) | Highest health score (`healthy` > `degraded` > `unknown` > `unhealthy`) | Unhealthy keys skipped | **PASS** |
| **Lowest Error Rate** (`lowest_error_rate`) | Smallest error-to-request ratio | Lower error rate key selected | **PASS** |

---

## 2. Cooldown & Exclusion Rules

A candidate is strictly excluded from rotator consideration if:
1. `Status != "active"` (manually disabled or suspended).
2. `ExpiresAt` is non-nil and in the past.
3. `CooldownUntil` is non-nil and in the future.
4. `HealthState == HealthUnhealthy` or `HealthExpired` (when healthy candidates exist).

---

## 3. High-Concurrency Stress Test

In `TestConcurrentRotatorSafety`:
- 20 concurrent goroutines executed 50 rotation operations each (1,000 total operations) on shared pool cursor.
- Atomic cursor increment (`atomic.AddUint64`) verified:
  - 0 deadlocks observed.
  - 0 panic conditions.
  - 0 index out of bounds.
  - Equal distribution across active credentials.

---

## 4. Pool Exhaustion Behavior

When all credentials in a pool are either disabled or in active cooldown:
- `rot.Select` immediately returns `error: no eligible credential available`.
- No infinite loops or thread blocking occurs.
- The HTTP layer converts this into a structured 502 Bad Gateway response.
