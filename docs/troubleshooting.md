# Troubleshooting Guide

## 1. Database Locked Error (SQLite)
- **Symptom:** `database is locked (5)`.
- **Cause:** Multiple write transactions contending without WAL mode or long-running exclusive locks.
- **Resolution:** EkaRouter automatically enables `journal_mode=WAL` and `busy_timeout=5000`. Ensure that no external SQLite tool (like DB Browser) has opened the database file with an exclusive lock.

## 2. SSRF Violation Error
- **Symptom:** `SSRF rejection: target host 127.0.0.1 is not permitted`.
- **Cause:** Outbound HTTP request attempted to reach a loopback, RFC 1918 private, or metadata IP.
- **Resolution:** If testing against local test servers on a developer machine, launch EkaRouter with `EKAROUTER_ALLOW_LOCAL_PROVIDERS=true`.

## 3. Unauthorized Error on Platform Endpoints
- **Symptom:** `{"success":false,"error":{"code":"unauthorized","message":"invalid or expired credentials"}}`.
- **Cause:** Missing or invalid `Authorization: Bearer <token>` header.
- **Resolution:** Provide an authenticated admin session token, a client token (`eka_pat_*`), or an API key (`eka_live_*`).
