# OAuth Package

## Purpose
Manages OAuth 2.0 PKCE states, cryptographic state hashing, single-use state consumption, and concurrency locks against token refresh stampedes.

## Files
- `oauth.go`: `StateRecord`, `Manager` with PKCE generation and singleflight account-level refresh mutex locks.
- `oauth_test.go`: Unit tests for single-use state consumption, TTL expiry, and concurrent refresh mutual exclusion.

## Allowed Responsibilities
- Ephemeral state generation, hashing, and validation.
- Mutual exclusion per account during credential refresh.

## Forbidden Responsibilities
- Storing decrypted tokens or plaintext secrets.
