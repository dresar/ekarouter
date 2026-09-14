# EkaRouter Final QA Test Plan

**Scope:** Backend Release Readiness & Pre-Frontend Verification  
**Audience:** QA Leads, Release Engineers, Frontend Engineers  
**Date:** 2026-09-15

---

## 1. Test Strategy & Objectives

The primary objective of this testing campaign is to guarantee that EkaRouter backend services are completely stable, secure, documented, and resilient before frontend interface implementation commences.

### Testing Principles
1. **Zero Fake Results**: Every reported pass or limitation is proven through automated code or test assertions.
2. **Credential Protection**: No cleartext secrets, passwords, or live tokens are recorded in logs, reports, or fixtures.
3. **SSRF Neutralization**: Outbound HTTP traffic is validated against loopback, link-local, private, and CGNAT IP blocks.
4. **Resilience & Fallback**: Routing fallbacks, rotation algorithms, and circuit breakers must prevent cascading failures.

---

## 2. Test Tiers & Execution Matrix

| Tier | Category | Coverage | Directory | Runner Command |
|---|---|---|---|---|
| **Tier 1** | Unit Tests | Data structures, algorithms, auth crypto, tokensaver, rotator | `internal/...` | `go test ./internal/...` |
| **Tier 2** | Integration Tests | HTTP handlers, middleware, database transactions, vault | `tests/integration/` | `go test ./tests/integration/...` |
| **Tier 3** | Security Tests | Secret masking, SSRF validation, body limits, header injection | `tests/integration/` | `go test ./tests/integration -run TestSecurity` |
| **Tier 4** | End-to-End Tests | Full user lifecycle from admin login to model inference | `tests/e2e/` | `go test ./tests/e2e/...` |
| **Tier 5** | Live Provider Mode | Real provider endpoint validation (Opt-In only) | `tests/integration/` | `EKAROUTER_ENABLE_LIVE_PROVIDER_TESTS=true go test` |

---

## 3. Dedicated Test Environment

- **Database**: Isolated temporary SQLite database (`data/test_*.db`) initialized with all 4 migration scripts.
- **Crypto Master Key**: Dedicated 32-character test key (`very-strong-secret-key-32-chars-long`).
- **Sandbox**: Local provider simulation enabled (`EKAROUTER_ALLOW_LOCAL_PROVIDERS=true`) with MockUpstream servers.
- **Cleanup**: Automatic teardown of database connections and temporary fixtures upon completion.

---

## 4. Acceptance Criteria for Release

1. All 55 internal packages and test suites compile and pass.
2. Integration tests verify system probes, session auth, vault credentials, rotation policies, and proxy CRUD.
3. Secret scanning confirms that no cleartext API key or proxy password is ever returned in JSON responses or log streams.
4. Windows binary compilation succeeds producing a standalone executable (`ekarouter.exe`).
5. Complete frontend contract and OpenAPI 3.1 specifications are generated and synchronized.
