# Auth Package

## Purpose
Provides symmetric AES-256-GCM authenticated encryption for upstream credentials at rest, SHA-256 one-way hashing for API keys and dashboard sessions, and cryptographically secure token generators.

## Files
- `auth.go`: `CryptoService` implementation with AES-GCM, `HashToken`, `GenerateApiKey`, and `GenerateSessionToken`.
- `auth_test.go`: Unit tests for encryption, decryption, hashing, and token generator formats.

## Allowed Responsibilities
- Secret encryption and decryption using a derived SHA-256 master key.
- Generation of high-entropy API keys and session tokens.
- Constant-time and secure token hashing.

## Forbidden Responsibilities
- No raw token logging or plaintext credential exposure.
- No direct HTTP routing.
